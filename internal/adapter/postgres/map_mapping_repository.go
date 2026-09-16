package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hackz-mosa-back/internal/domain"
)

const mapMappingSchema = `
CREATE TABLE IF NOT EXISTS map_mappings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    area JSONB NOT NULL,
    track_id TEXT NOT NULL,
    title TEXT NOT NULL,
    artist TEXT NOT NULL,
    artwork_url TEXT,
    audius_url TEXT,
    duration_seconds INTEGER CHECK (duration_seconds IS NULL OR duration_seconds >= 0),
    stream_url TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source = 'audius'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS map_mappings_user_id_created_at_idx
    ON map_mappings (user_id, created_at DESC, id DESC);
`

type MapMappingRepository struct {
	pool *pgxpool.Pool
}

func NewMapMappingRepository(users *UserRepository) *MapMappingRepository {
	return &MapMappingRepository{pool: users.pool}
}

func (r *MapMappingRepository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, mapMappingSchema); err != nil {
		return fmt.Errorf("migrate map mappings schema: %w", err)
	}
	return nil
}

func (r *MapMappingRepository) Create(ctx context.Context, ownerID int64, mapping domain.NewMapMapping) (domain.MapMapping, error) {
	areaJSON, err := json.Marshal(mapping.Area)
	if err != nil {
		return domain.MapMapping{}, fmt.Errorf("encode map mapping area: %w", err)
	}

	const query = `
INSERT INTO map_mappings (
    user_id, area, track_id, title, artist, artwork_url, audius_url,
    duration_seconds, stream_url, source
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, created_at
`
	created := domain.MapMapping{Area: mapping.Area, Track: mapping.Track}
	err = r.pool.QueryRow(ctx, query,
		ownerID,
		string(areaJSON),
		mapping.Track.TrackID,
		mapping.Track.Title,
		mapping.Track.Artist,
		mapping.Track.ArtworkURL,
		mapping.Track.AudiusURL,
		mapping.Track.DurationSeconds,
		mapping.Track.StreamURL,
		mapping.Track.Source,
	).Scan(&created.ID, &created.CreatedAt)
	if err != nil {
		return domain.MapMapping{}, fmt.Errorf("create map mapping: %w", err)
	}
	return created, nil
}

func (r *MapMappingRepository) List(ctx context.Context, ownerID int64) ([]domain.MapMapping, error) {
	const query = `
SELECT id, area, track_id, title, artist, artwork_url, audius_url,
       duration_seconds, stream_url, source, created_at
FROM map_mappings
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list map mappings: %w", err)
	}
	mappings, err := pgx.CollectRows(rows, scanMapMapping)
	if err != nil {
		return nil, fmt.Errorf("scan map mappings: %w", err)
	}
	return mappings, nil
}

func (r *MapMappingRepository) Delete(ctx context.Context, ownerID, mappingID int64) error {
	command, err := r.pool.Exec(ctx, `DELETE FROM map_mappings WHERE id = $1 AND user_id = $2`, mappingID, ownerID)
	if err != nil {
		return fmt.Errorf("delete map mapping: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ErrMapMappingNotFound
	}
	return nil
}

func scanMapMapping(row pgx.CollectableRow) (domain.MapMapping, error) {
	var mapping domain.MapMapping
	var areaJSON []byte
	err := row.Scan(
		&mapping.ID,
		&areaJSON,
		&mapping.Track.TrackID,
		&mapping.Track.Title,
		&mapping.Track.Artist,
		&mapping.Track.ArtworkURL,
		&mapping.Track.AudiusURL,
		&mapping.Track.DurationSeconds,
		&mapping.Track.StreamURL,
		&mapping.Track.Source,
		&mapping.CreatedAt,
	)
	if err != nil {
		return domain.MapMapping{}, err
	}
	if err := json.Unmarshal(areaJSON, &mapping.Area); err != nil {
		return domain.MapMapping{}, fmt.Errorf("decode map mapping area: %w", err)
	}
	if len(mapping.Area) == 0 {
		return domain.MapMapping{}, errors.New("map mapping area is empty")
	}
	return mapping, nil
}
