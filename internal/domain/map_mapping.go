package domain

import (
	"context"
	"errors"
	"time"
)

const MaxMappingAreaPoints = 500

var (
	ErrInvalidMapMapping  = errors.New("invalid map mapping")
	ErrMapMappingNotFound = errors.New("map mapping not found")
)

type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type MappedTrack struct {
	TrackID         string  `json:"id"`
	Title           string  `json:"title"`
	Artist          string  `json:"artist"`
	ArtworkURL      *string `json:"artworkUrl,omitempty"`
	AudiusURL       *string `json:"audiusUrl,omitempty"`
	DurationSeconds *int    `json:"durationSeconds,omitempty"`
	StreamURL       string  `json:"streamUrl"`
	Source          string  `json:"source"`
}

type MapMapping struct {
	ID        int64       `json:"id"`
	Area      []GeoPoint  `json:"area"`
	Track     MappedTrack `json:"track"`
	CreatedAt time.Time   `json:"createdAt"`
}

type NewMapMapping struct {
	Area  []GeoPoint
	Track MappedTrack
}

type MapMappingRepository interface {
	Create(ctx context.Context, ownerID int64, mapping NewMapMapping) (MapMapping, error)
	List(ctx context.Context, ownerID int64) ([]MapMapping, error)
	Delete(ctx context.Context, ownerID, mappingID int64) error
}
