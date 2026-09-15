package domain

import (
	"context"
	"errors"
	"time"
)

const MaxPlaylistTracks = 30

var (
	ErrInvalidPlaylist   = errors.New("invalid playlist")
	ErrPlaylistNotFound  = errors.New("playlist not found")
	ErrDuplicateTrack    = errors.New("track is already in playlist")
	ErrPlaylistFull      = errors.New("playlist has reached its track limit")
	ErrInvalidTrackOrder = errors.New("track order must contain every playlist item exactly once")
)

type PlaylistTrack struct {
	ID              int64     `json:"id"`
	TrackID         string    `json:"trackId"`
	Title           string    `json:"title"`
	Artist          string    `json:"artist"`
	ArtworkURL      *string   `json:"artworkUrl,omitempty"`
	AudiusURL       *string   `json:"audiusUrl,omitempty"`
	DurationSeconds *int      `json:"durationSeconds,omitempty"`
	StreamURL       string    `json:"streamUrl"`
	Source          string    `json:"source"`
	Position        int       `json:"position"`
	CreatedAt       time.Time `json:"createdAt"`
}

type NewPlaylistTrack struct {
	TrackID         string  `json:"trackId"`
	Title           string  `json:"title"`
	Artist          string  `json:"artist"`
	ArtworkURL      *string `json:"artworkUrl,omitempty"`
	AudiusURL       *string `json:"audiusUrl,omitempty"`
	DurationSeconds *int    `json:"durationSeconds,omitempty"`
	StreamURL       string  `json:"streamUrl"`
	Source          string  `json:"source"`
}

type Playlist struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	TrackCount  int             `json:"trackCount"`
	Tracks      []PlaylistTrack `json:"tracks"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

type PlaylistRepository interface {
	Create(ctx context.Context, ownerID int64, name, description string) (Playlist, error)
	List(ctx context.Context, ownerID int64) ([]Playlist, error)
	Get(ctx context.Context, ownerID, playlistID int64) (Playlist, error)
	Delete(ctx context.Context, ownerID, playlistID int64) error
	AddTrack(ctx context.Context, ownerID, playlistID int64, track NewPlaylistTrack) (Playlist, error)
	ReorderTracks(ctx context.Context, ownerID, playlistID int64, itemIDs []int64) (Playlist, error)
	DeleteTrack(ctx context.Context, ownerID, playlistID, itemID int64) (Playlist, error)
}
