package usecase

import (
	"context"
	"errors"
	"testing"

	"hackz-mosa-back/internal/domain"
)

type mapMappingRepositoryStub struct {
	created domain.NewMapMapping
}

func (s *mapMappingRepositoryStub) Create(_ context.Context, _ int64, mapping domain.NewMapMapping) (domain.MapMapping, error) {
	s.created = mapping
	return domain.MapMapping{ID: 1, Area: mapping.Area, Track: mapping.Track}, nil
}

func (s *mapMappingRepositoryStub) List(context.Context, int64) ([]domain.MapMapping, error) {
	return []domain.MapMapping{}, nil
}

func (s *mapMappingRepositoryStub) Delete(context.Context, int64, int64) error {
	return nil
}

func TestMapMappingServiceCreateNormalizesClosedArea(t *testing.T) {
	repository := &mapMappingRepositoryStub{}
	service := NewMapMappingService(repository)
	first := domain.GeoPoint{Latitude: 35.68, Longitude: 139.76}

	_, err := service.Create(context.Background(), 1, domain.NewMapMapping{
		Area: []domain.GeoPoint{
			first,
			{Latitude: 35.69, Longitude: 139.76},
			{Latitude: 35.69, Longitude: 139.77},
			first,
		},
		Track: validMappedTrack(),
	})

	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(repository.created.Area) != 3 {
		t.Fatalf("area length = %d, want 3", len(repository.created.Area))
	}
}

func TestMapMappingServiceCreateRejectsInvalidArea(t *testing.T) {
	service := NewMapMappingService(&mapMappingRepositoryStub{})

	_, err := service.Create(context.Background(), 1, domain.NewMapMapping{
		Area: []domain.GeoPoint{
			{Latitude: 35.68, Longitude: 139.76},
			{Latitude: 91, Longitude: 139.77},
		},
		Track: validMappedTrack(),
	})

	if !errors.Is(err, domain.ErrInvalidMapMapping) {
		t.Fatalf("Create() error = %v, want ErrInvalidMapMapping", err)
	}
}

func TestMapMappingServiceCreateRejectsNonAudiusTrack(t *testing.T) {
	service := NewMapMappingService(&mapMappingRepositoryStub{})
	track := validMappedTrack()
	track.Source = "other"

	_, err := service.Create(context.Background(), 1, domain.NewMapMapping{
		Area: []domain.GeoPoint{
			{Latitude: 35.68, Longitude: 139.76},
			{Latitude: 35.69, Longitude: 139.76},
			{Latitude: 35.69, Longitude: 139.77},
		},
		Track: track,
	})

	if !errors.Is(err, domain.ErrInvalidMapMapping) {
		t.Fatalf("Create() error = %v, want ErrInvalidMapMapping", err)
	}
}

func validMappedTrack() domain.MappedTrack {
	return domain.MappedTrack{
		TrackID:   "track-1",
		Title:     "Mapped song",
		Artist:    "Artist",
		StreamURL: "https://api.audius.co/v1/tracks/track-1/stream",
		Source:    "audius",
	}
}
