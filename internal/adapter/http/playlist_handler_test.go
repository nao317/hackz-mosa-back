package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"hackz-mosa-back/internal/domain"
)

type playlistUseCaseStub struct {
	ownerID          int64
	name             string
	track            domain.NewPlaylistTrack
	reorderedItemIDs []int64
	addTrackErr      error
}

func (s *playlistUseCaseStub) Create(_ context.Context, ownerID int64, name, _ string) (domain.Playlist, error) {
	s.ownerID = ownerID
	s.name = name
	return domain.Playlist{ID: 7, Name: name, Tracks: []domain.PlaylistTrack{}}, nil
}

func (s *playlistUseCaseStub) List(context.Context, int64) ([]domain.Playlist, error) {
	return []domain.Playlist{}, nil
}

func (s *playlistUseCaseStub) Get(context.Context, int64, int64) (domain.Playlist, error) {
	return domain.Playlist{}, nil
}

func (s *playlistUseCaseStub) Delete(context.Context, int64, int64) error {
	return nil
}

func (s *playlistUseCaseStub) AddTrack(_ context.Context, ownerID, _ int64, track domain.NewPlaylistTrack) (domain.Playlist, error) {
	s.ownerID = ownerID
	s.track = track
	return domain.Playlist{}, s.addTrackErr
}

func (s *playlistUseCaseStub) ReorderTracks(_ context.Context, _, _ int64, itemIDs []int64) (domain.Playlist, error) {
	s.reorderedItemIDs = itemIDs
	return domain.Playlist{}, nil
}

func (s *playlistUseCaseStub) DeleteTrack(context.Context, int64, int64, int64) (domain.Playlist, error) {
	return domain.Playlist{}, nil
}

func playlistTestApp(service *playlistUseCaseStub) *echo.Echo {
	app := echo.New()
	authenticate := &signInStub{}
	NewPlaylistHandler(authenticate, service).Register(app.Group("/api/v1"))
	return app
}

func TestCreatePlaylist(t *testing.T) {
	service := &playlistUseCaseStub{}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/playlists", strings.NewReader(`{"name":"朝の曲","description":"通勤用"}`))
	request.Header.Set("Authorization", "Bearer firebase-id-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	playlistTestApp(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if service.ownerID != 42 || service.name != "朝の曲" {
		t.Fatalf("ownerID = %d, name = %q", service.ownerID, service.name)
	}
}

func TestAddPlaylistTrackUsesFrontendTrackShape(t *testing.T) {
	service := &playlistUseCaseStub{}
	body := `{"id":"track-1","title":"Song","artist":"Artist","streamUrl":"https://api.audius.co/v1/tracks/track-1/stream","source":"audius"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/playlists/7/tracks", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer firebase-id-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	playlistTestApp(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if service.track.TrackID != "track-1" || service.track.Source != "audius" {
		t.Fatalf("track = %#v", service.track)
	}
}

func TestAddPlaylistTrackReturnsLimitError(t *testing.T) {
	service := &playlistUseCaseStub{addTrackErr: domain.ErrPlaylistFull}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/playlists/7/tracks", strings.NewReader(`{"id":"track-31"}`))
	request.Header.Set("Authorization", "Bearer firebase-id-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	playlistTestApp(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), "playlist_full") {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestReorderPlaylistTracks(t *testing.T) {
	service := &playlistUseCaseStub{}
	request := httptest.NewRequest(http.MethodPut, "/api/v1/playlists/7/tracks/order", strings.NewReader(`{"itemIds":[3,1,2]}`))
	request.Header.Set("Authorization", "Bearer firebase-id-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	playlistTestApp(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if got := service.reorderedItemIDs; len(got) != 3 || got[0] != 3 || got[2] != 2 {
		t.Fatalf("item IDs = %#v", got)
	}
}

func TestPlaylistRoutesRequireAuthentication(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/playlists", nil)
	recorder := httptest.NewRecorder()

	playlistTestApp(&playlistUseCaseStub{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
