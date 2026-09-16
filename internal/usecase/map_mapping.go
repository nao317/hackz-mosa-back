package usecase

import (
	"context"
	"math"
	"strings"

	"hackz-mosa-back/internal/domain"
)

type MapMappingService struct {
	mappings domain.MapMappingRepository
}

func NewMapMappingService(mappings domain.MapMappingRepository) *MapMappingService {
	return &MapMappingService{mappings: mappings}
}

func (s *MapMappingService) Create(ctx context.Context, ownerID int64, mapping domain.NewMapMapping) (domain.MapMapping, error) {
	mapping.Track.TrackID = strings.TrimSpace(mapping.Track.TrackID)
	mapping.Track.Title = strings.TrimSpace(mapping.Track.Title)
	mapping.Track.Artist = strings.TrimSpace(mapping.Track.Artist)
	mapping.Track.StreamURL = strings.TrimSpace(mapping.Track.StreamURL)
	mapping.Track.Source = strings.TrimSpace(mapping.Track.Source)
	mapping.Area = normalizeArea(mapping.Area)

	if ownerID <= 0 || len(mapping.Area) < 3 || len(mapping.Area) > domain.MaxMappingAreaPoints {
		return domain.MapMapping{}, domain.ErrInvalidMapMapping
	}
	for _, point := range mapping.Area {
		if !validCoordinate(point) {
			return domain.MapMapping{}, domain.ErrInvalidMapMapping
		}
	}
	if mapping.Track.TrackID == "" || mapping.Track.Title == "" || mapping.Track.Artist == "" ||
		mapping.Track.StreamURL == "" || mapping.Track.Source != "audius" {
		return domain.MapMapping{}, domain.ErrInvalidMapMapping
	}
	if mapping.Track.DurationSeconds != nil && *mapping.Track.DurationSeconds < 0 {
		return domain.MapMapping{}, domain.ErrInvalidMapMapping
	}

	return s.mappings.Create(ctx, ownerID, mapping)
}

func (s *MapMappingService) List(ctx context.Context, ownerID int64) ([]domain.MapMapping, error) {
	if ownerID <= 0 {
		return nil, domain.ErrInvalidMapMapping
	}
	return s.mappings.List(ctx, ownerID)
}

func (s *MapMappingService) Delete(ctx context.Context, ownerID, mappingID int64) error {
	if ownerID <= 0 || mappingID <= 0 {
		return domain.ErrMapMappingNotFound
	}
	return s.mappings.Delete(ctx, ownerID, mappingID)
}

func normalizeArea(area []domain.GeoPoint) []domain.GeoPoint {
	if len(area) > 1 && area[0] == area[len(area)-1] {
		area = area[:len(area)-1]
	}
	return area
}

func validCoordinate(point domain.GeoPoint) bool {
	return !math.IsNaN(point.Latitude) && !math.IsInf(point.Latitude, 0) &&
		!math.IsNaN(point.Longitude) && !math.IsInf(point.Longitude, 0) &&
		point.Latitude >= -90 && point.Latitude <= 90 &&
		point.Longitude >= -180 && point.Longitude <= 180
}
