package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"hackz-mosa-back/internal/domain"
)

const playlistSchema = `
CREATE TABLE IF NOT EXISTS playlists (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
    description TEXT NOT NULL DEFAULT '' CHECK (char_length(description) <= 500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS playlists_user_id_updated_at_idx
    ON playlists (user_id, updated_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS playlist_tracks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    playlist_id BIGINT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id TEXT NOT NULL,
    title TEXT NOT NULL,
    artist TEXT NOT NULL,
    artwork_url TEXT,
    audius_url TEXT,
    duration_seconds INTEGER CHECK (duration_seconds IS NULL OR duration_seconds >= 0),
    stream_url TEXT NOT NULL,
    source TEXT NOT NULL,
    position INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (playlist_id, track_id),
    UNIQUE (playlist_id, position)
);

CREATE INDEX IF NOT EXISTS playlist_tracks_playlist_id_position_idx
    ON playlist_tracks (playlist_id, position);
`

type PlaylistRepository struct {
	pool *pgxpool.Pool
}

func NewPlaylistRepository(users *UserRepository) *PlaylistRepository {
	return &PlaylistRepository{pool: users.pool}
}

func (r *PlaylistRepository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, playlistSchema); err != nil {
		return fmt.Errorf("migrate playlists schema: %w", err)
	}
	return nil
}

func (r *PlaylistRepository) Create(ctx context.Context, ownerID int64, name, description string) (domain.Playlist, error) {
	const query = `
INSERT INTO playlists (user_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, name, description, created_at, updated_at
`
	playlist := domain.Playlist{Tracks: []domain.PlaylistTrack{}}
	err := r.pool.QueryRow(ctx, query, ownerID, name, description).Scan(
		&playlist.ID,
		&playlist.Name,
		&playlist.Description,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
	)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("create playlist: %w", err)
	}
	return playlist, nil
}

func (r *PlaylistRepository) List(ctx context.Context, ownerID int64) ([]domain.Playlist, error) {
	const query = `
SELECT id, name, description, created_at, updated_at
FROM playlists
WHERE user_id = $1
ORDER BY updated_at DESC, id DESC
`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list playlists: %w", err)
	}
	playlists, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Playlist, error) {
		playlist := domain.Playlist{Tracks: []domain.PlaylistTrack{}}
		err := row.Scan(&playlist.ID, &playlist.Name, &playlist.Description, &playlist.CreatedAt, &playlist.UpdatedAt)
		return playlist, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan playlists: %w", err)
	}
	for index := range playlists {
		tracks, err := loadPlaylistTracks(ctx, r.pool, playlists[index].ID)
		if err != nil {
			return nil, err
		}
		playlists[index].Tracks = tracks
		playlists[index].TrackCount = len(tracks)
	}
	return playlists, nil
}

func (r *PlaylistRepository) Get(ctx context.Context, ownerID, playlistID int64) (domain.Playlist, error) {
	return loadPlaylist(ctx, r.pool, ownerID, playlistID)
}

func (r *PlaylistRepository) Delete(ctx context.Context, ownerID, playlistID int64) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM playlists WHERE id = $1 AND user_id = $2`, playlistID, ownerID)
	if err != nil {
		return fmt.Errorf("delete playlist: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ErrPlaylistNotFound
	}
	return nil
}

func (r *PlaylistRepository) AddTrack(ctx context.Context, ownerID, playlistID int64, track domain.NewPlaylistTrack) (domain.Playlist, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("begin add track transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := lockOwnedPlaylist(ctx, tx, ownerID, playlistID); err != nil {
		return domain.Playlist{}, err
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM playlist_tracks WHERE playlist_id = $1`, playlistID).Scan(&count); err != nil {
		return domain.Playlist{}, fmt.Errorf("count playlist tracks: %w", err)
	}
	if count >= domain.MaxPlaylistTracks {
		return domain.Playlist{}, domain.ErrPlaylistFull
	}

	const insert = `
INSERT INTO playlist_tracks (
    playlist_id, track_id, title, artist, artwork_url, audius_url,
    duration_seconds, stream_url, source, position
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`
	_, err = tx.Exec(ctx, insert,
		playlistID,
		track.TrackID,
		track.Title,
		track.Artist,
		track.ArtworkURL,
		track.AudiusURL,
		track.DurationSeconds,
		track.StreamURL,
		track.Source,
		count+1,
	)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return domain.Playlist{}, domain.ErrDuplicateTrack
		}
		return domain.Playlist{}, fmt.Errorf("add playlist track: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE playlists SET updated_at = NOW() WHERE id = $1`, playlistID); err != nil {
		return domain.Playlist{}, fmt.Errorf("touch playlist: %w", err)
	}
	playlist, err := loadPlaylist(ctx, tx, ownerID, playlistID)
	if err != nil {
		return domain.Playlist{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Playlist{}, fmt.Errorf("commit add track transaction: %w", err)
	}
	return playlist, nil
}

func (r *PlaylistRepository) ReorderTracks(ctx context.Context, ownerID, playlistID int64, itemIDs []int64) (domain.Playlist, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("begin reorder transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := lockOwnedPlaylist(ctx, tx, ownerID, playlistID); err != nil {
		return domain.Playlist{}, err
	}
	rows, err := tx.Query(ctx, `SELECT id FROM playlist_tracks WHERE playlist_id = $1 ORDER BY position FOR UPDATE`, playlistID)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("lock playlist tracks: %w", err)
	}
	existingIDs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (int64, error) {
		var id int64
		err := row.Scan(&id)
		return id, err
	})
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("scan playlist track ids: %w", err)
	}
	if !sameIDs(existingIDs, itemIDs) {
		return domain.Playlist{}, domain.ErrInvalidTrackOrder
	}

	if len(itemIDs) > 0 {
		if _, err := tx.Exec(ctx, `UPDATE playlist_tracks SET position = position + 100 WHERE playlist_id = $1`, playlistID); err != nil {
			return domain.Playlist{}, fmt.Errorf("prepare playlist reorder: %w", err)
		}
		for index, itemID := range itemIDs {
			if _, err := tx.Exec(ctx, `UPDATE playlist_tracks SET position = $1 WHERE id = $2 AND playlist_id = $3`, index+1, itemID, playlistID); err != nil {
				return domain.Playlist{}, fmt.Errorf("reorder playlist track: %w", err)
			}
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE playlists SET updated_at = NOW() WHERE id = $1`, playlistID); err != nil {
		return domain.Playlist{}, fmt.Errorf("touch playlist: %w", err)
	}
	playlist, err := loadPlaylist(ctx, tx, ownerID, playlistID)
	if err != nil {
		return domain.Playlist{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Playlist{}, fmt.Errorf("commit reorder transaction: %w", err)
	}
	return playlist, nil
}

func (r *PlaylistRepository) DeleteTrack(ctx context.Context, ownerID, playlistID, itemID int64) (domain.Playlist, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("begin delete track transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := lockOwnedPlaylist(ctx, tx, ownerID, playlistID); err != nil {
		return domain.Playlist{}, err
	}
	var position int
	if err := tx.QueryRow(ctx, `DELETE FROM playlist_tracks WHERE id = $1 AND playlist_id = $2 RETURNING position`, itemID, playlistID).Scan(&position); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Playlist{}, domain.ErrPlaylistNotFound
		}
		return domain.Playlist{}, fmt.Errorf("delete playlist track: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE playlist_tracks SET position = position + 100 WHERE playlist_id = $1 AND position > $2`, playlistID, position); err != nil {
		return domain.Playlist{}, fmt.Errorf("prepare playlist position compaction: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE playlist_tracks SET position = position - 101 WHERE playlist_id = $1 AND position > 100`, playlistID); err != nil {
		return domain.Playlist{}, fmt.Errorf("compact playlist positions: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE playlists SET updated_at = NOW() WHERE id = $1`, playlistID); err != nil {
		return domain.Playlist{}, fmt.Errorf("touch playlist: %w", err)
	}
	playlist, err := loadPlaylist(ctx, tx, ownerID, playlistID)
	if err != nil {
		return domain.Playlist{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Playlist{}, fmt.Errorf("commit delete track transaction: %w", err)
	}
	return playlist, nil
}

type playlistQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func loadPlaylist(ctx context.Context, queryer playlistQuerier, ownerID, playlistID int64) (domain.Playlist, error) {
	const query = `
SELECT id, name, description, created_at, updated_at
FROM playlists
WHERE id = $1 AND user_id = $2
`
	playlist := domain.Playlist{Tracks: []domain.PlaylistTrack{}}
	err := queryer.QueryRow(ctx, query, playlistID, ownerID).Scan(
		&playlist.ID,
		&playlist.Name,
		&playlist.Description,
		&playlist.CreatedAt,
		&playlist.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Playlist{}, domain.ErrPlaylistNotFound
	}
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("get playlist: %w", err)
	}
	playlist.Tracks, err = loadPlaylistTracks(ctx, queryer, playlistID)
	if err != nil {
		return domain.Playlist{}, err
	}
	playlist.TrackCount = len(playlist.Tracks)
	return playlist, nil
}

func loadPlaylistTracks(ctx context.Context, queryer playlistQuerier, playlistID int64) ([]domain.PlaylistTrack, error) {
	const query = `
SELECT id, track_id, title, artist, artwork_url, audius_url,
       duration_seconds, stream_url, source, position, created_at
FROM playlist_tracks
WHERE playlist_id = $1
ORDER BY position
`
	rows, err := queryer.Query(ctx, query, playlistID)
	if err != nil {
		return nil, fmt.Errorf("list playlist tracks: %w", err)
	}
	tracks, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.PlaylistTrack, error) {
		var track domain.PlaylistTrack
		err := row.Scan(
			&track.ID,
			&track.TrackID,
			&track.Title,
			&track.Artist,
			&track.ArtworkURL,
			&track.AudiusURL,
			&track.DurationSeconds,
			&track.StreamURL,
			&track.Source,
			&track.Position,
			&track.CreatedAt,
		)
		return track, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan playlist tracks: %w", err)
	}
	return tracks, nil
}

func lockOwnedPlaylist(ctx context.Context, tx pgx.Tx, ownerID, playlistID int64) error {
	var id int64
	if err := tx.QueryRow(ctx, `SELECT id FROM playlists WHERE id = $1 AND user_id = $2 FOR UPDATE`, playlistID, ownerID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrPlaylistNotFound
		}
		return fmt.Errorf("lock playlist: %w", err)
	}
	return nil
}

func sameIDs(existing, requested []int64) bool {
	if len(existing) != len(requested) {
		return false
	}
	ids := make(map[int64]struct{}, len(existing))
	for _, id := range existing {
		ids[id] = struct{}{}
	}
	for _, id := range requested {
		if _, exists := ids[id]; !exists {
			return false
		}
		delete(ids, id)
	}
	return len(ids) == 0
}
