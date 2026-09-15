package usecase

import (
	"context"
	"errors"
	"testing"

	"hackz-mosa-back/internal/domain"
)

type playlistRepositoryStub struct {
	createdName        string
	createdDescription string
	addedTrack         domain.NewPlaylistTrack
	reorderedItemIDs   []int64
}

func (s *playlistRepositoryStub) Create(_ context.Context, _ int64, name, description string) (domain.Playlist, error) {
	s.createdName = name
	s.createdDescription = description
	return domain.Playlist{Name: name, Description: description}, nil
}

func (s *playlistRepositoryStub) List(context.Context, int64) ([]domain.Playlist, error) {
	return []domain.Playlist{}, nil
}

func (s *playlistRepositoryStub) Get(context.Context, int64, int64) (domain.Playlist, error) {
	return domain.Playlist{}, nil
}

func (s *playlistRepositoryStub) Delete(context.Context, int64, int64) error {
	return nil
}

func (s *playlistRepositoryStub) AddTrack(_ context.Context, _, _ int64, track domain.NewPlaylistTrack) (domain.Playlist, error) {
	s.addedTrack = track
	return domain.Playlist{}, nil
}

func (s *playlistRepositoryStub) ReorderTracks(_ context.Context, _, _ int64, itemIDs []int64) (domain.Playlist, error) {
	s.reorderedItemIDs = itemIDs
	return domain.Playlist{}, nil
}

func (s *playlistRepositoryStub) DeleteTrack(context.Context, int64, int64, int64) (domain.Playlist, error) {
	return domain.Playlist{}, nil
}

func TestPlaylistServiceCreateNormalizesText(t *testing.T) {
	repository := &playlistRepositoryStub{}
	service := NewPlaylistService(repository)

	_, err := service.Create(context.Background(), 1, "  朝の曲  ", "  通勤用  ")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.createdName != "朝の曲" || repository.createdDescription != "通勤用" {
		t.Fatalf("created playlist = %q, %q", repository.createdName, repository.createdDescription)
	}
}

func TestPlaylistServiceCreateRejectsEmptyName(t *testing.T) {
	service := NewPlaylistService(&playlistRepositoryStub{})

	_, err := service.Create(context.Background(), 1, "  ", "description")

	if !errors.Is(err, domain.ErrInvalidPlaylist) {
		t.Fatalf("Create() error = %v, want ErrInvalidPlaylist", err)
	}
}

func TestPlaylistServiceAddTrackAcceptsFrontendTrack(t *testing.T) {
	repository := &playlistRepositoryStub{}
	service := NewPlaylistService(repository)

	_, err := service.AddTrack(context.Background(), 1, 2, domain.NewPlaylistTrack{
		TrackID:   " audius-track ",
		Title:     " Title ",
		Artist:    " Artist ",
		StreamURL: " https://api.audius.co/v1/tracks/audius-track/stream ",
		Source:    " audius ",
	})
	if err != nil {
		t.Fatalf("AddTrack() error = %v", err)
	}
	if repository.addedTrack.TrackID != "audius-track" || repository.addedTrack.Title != "Title" {
		t.Fatalf("added track = %#v", repository.addedTrack)
	}
}

func TestPlaylistServiceReorderRejectsDuplicateItemIDs(t *testing.T) {
	service := NewPlaylistService(&playlistRepositoryStub{})

	_, err := service.ReorderTracks(context.Background(), 1, 2, []int64{10, 10})

	if !errors.Is(err, domain.ErrInvalidTrackOrder) {
		t.Fatalf("ReorderTracks() error = %v, want ErrInvalidTrackOrder", err)
	}
}

func TestPlaylistServiceReorderRejectsMoreThanThirtyItems(t *testing.T) {
	service := NewPlaylistService(&playlistRepositoryStub{})
	itemIDs := make([]int64, domain.MaxPlaylistTracks+1)
	for index := range itemIDs {
		itemIDs[index] = int64(index + 1)
	}

	_, err := service.ReorderTracks(context.Background(), 1, 2, itemIDs)

	if !errors.Is(err, domain.ErrInvalidTrackOrder) {
		t.Fatalf("ReorderTracks() error = %v, want ErrInvalidTrackOrder", err)
	}
}
