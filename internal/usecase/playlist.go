package usecase

import (
	"context"
	"fmt"
	"strings"

	"hackz-mosa-back/internal/domain"
)

const (
	maxPlaylistNameLength        = 100
	maxPlaylistDescriptionLength = 500
)

type PlaylistService struct {
	playlists domain.PlaylistRepository
}

func NewPlaylistService(playlists domain.PlaylistRepository) *PlaylistService {
	return &PlaylistService{playlists: playlists}
}

func (s *PlaylistService) Create(ctx context.Context, ownerID int64, name, description string) (domain.Playlist, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if ownerID <= 0 || name == "" || len([]rune(name)) > maxPlaylistNameLength || len([]rune(description)) > maxPlaylistDescriptionLength {
		return domain.Playlist{}, domain.ErrInvalidPlaylist
	}
	return s.playlists.Create(ctx, ownerID, name, description)
}

func (s *PlaylistService) List(ctx context.Context, ownerID int64) ([]domain.Playlist, error) {
	if ownerID <= 0 {
		return nil, domain.ErrInvalidPlaylist
	}
	return s.playlists.List(ctx, ownerID)
}

func (s *PlaylistService) Get(ctx context.Context, ownerID, playlistID int64) (domain.Playlist, error) {
	if ownerID <= 0 || playlistID <= 0 {
		return domain.Playlist{}, domain.ErrPlaylistNotFound
	}
	return s.playlists.Get(ctx, ownerID, playlistID)
}

func (s *PlaylistService) Delete(ctx context.Context, ownerID, playlistID int64) error {
	if ownerID <= 0 || playlistID <= 0 {
		return domain.ErrPlaylistNotFound
	}
	return s.playlists.Delete(ctx, ownerID, playlistID)
}

func (s *PlaylistService) AddTrack(ctx context.Context, ownerID, playlistID int64, track domain.NewPlaylistTrack) (domain.Playlist, error) {
	track.TrackID = strings.TrimSpace(track.TrackID)
	track.Title = strings.TrimSpace(track.Title)
	track.Artist = strings.TrimSpace(track.Artist)
	track.StreamURL = strings.TrimSpace(track.StreamURL)
	track.Source = strings.TrimSpace(track.Source)
	if ownerID <= 0 || playlistID <= 0 || track.TrackID == "" || track.Title == "" || track.Artist == "" || track.StreamURL == "" || track.Source == "" {
		return domain.Playlist{}, domain.ErrInvalidPlaylist
	}
	if track.DurationSeconds != nil && *track.DurationSeconds < 0 {
		return domain.Playlist{}, domain.ErrInvalidPlaylist
	}
	return s.playlists.AddTrack(ctx, ownerID, playlistID, track)
}

func (s *PlaylistService) ReorderTracks(ctx context.Context, ownerID, playlistID int64, itemIDs []int64) (domain.Playlist, error) {
	if ownerID <= 0 || playlistID <= 0 || len(itemIDs) > domain.MaxPlaylistTracks {
		return domain.Playlist{}, domain.ErrInvalidTrackOrder
	}
	seen := make(map[int64]struct{}, len(itemIDs))
	for _, itemID := range itemIDs {
		if itemID <= 0 {
			return domain.Playlist{}, domain.ErrInvalidTrackOrder
		}
		if _, exists := seen[itemID]; exists {
			return domain.Playlist{}, domain.ErrInvalidTrackOrder
		}
		seen[itemID] = struct{}{}
	}
	return s.playlists.ReorderTracks(ctx, ownerID, playlistID, itemIDs)
}

func (s *PlaylistService) DeleteTrack(ctx context.Context, ownerID, playlistID, itemID int64) (domain.Playlist, error) {
	if ownerID <= 0 || playlistID <= 0 || itemID <= 0 {
		return domain.Playlist{}, domain.ErrPlaylistNotFound
	}
	playlist, err := s.playlists.DeleteTrack(ctx, ownerID, playlistID, itemID)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("delete playlist track: %w", err)
	}
	return playlist, nil
}
