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

type mapMappingUseCaseStub struct {
	ownerID  int64
	created  domain.NewMapMapping
	deleted  int64
	mappings []domain.MapMapping
}

func (s *mapMappingUseCaseStub) Create(_ context.Context, ownerID int64, mapping domain.NewMapMapping) (domain.MapMapping, error) {
	s.ownerID = ownerID
	s.created = mapping
	return domain.MapMapping{ID: 9, Area: mapping.Area, Track: mapping.Track}, nil
}

func (s *mapMappingUseCaseStub) List(_ context.Context, ownerID int64) ([]domain.MapMapping, error) {
	s.ownerID = ownerID
	return s.mappings, nil
}

func (s *mapMappingUseCaseStub) Delete(_ context.Context, ownerID, mappingID int64) error {
	s.ownerID = ownerID
	s.deleted = mappingID
	return nil
}

func mapMappingTestApp(service *mapMappingUseCaseStub) *echo.Echo {
	app := echo.New()
	NewMapMappingHandler(&signInStub{}, service).Register(app.Group("/api/v1"))
	return app
}

func TestCreateMapMapping(t *testing.T) {
	service := &mapMappingUseCaseStub{}
	body := `{
		"area":[
			{"latitude":35.68,"longitude":139.76},
			{"latitude":35.69,"longitude":139.76},
			{"latitude":35.69,"longitude":139.77}
		],
		"track":{"id":"track-1","title":"Song","artist":"Artist","streamUrl":"https://api.audius.co/v1/tracks/track-1/stream","source":"audius"}
	}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/map-mappings", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer firebase-id-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	mapMappingTestApp(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if service.ownerID != 42 || service.created.Track.TrackID != "track-1" || len(service.created.Area) != 3 {
		t.Fatalf("created mapping = %#v, ownerID = %d", service.created, service.ownerID)
	}
}

func TestDeleteMapMapping(t *testing.T) {
	service := &mapMappingUseCaseStub{}
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/map-mappings/9", nil)
	request.Header.Set("Authorization", "Bearer firebase-id-token")
	recorder := httptest.NewRecorder()

	mapMappingTestApp(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent || service.ownerID != 42 || service.deleted != 9 {
		t.Fatalf("status = %d, ownerID = %d, deleted = %d", recorder.Code, service.ownerID, service.deleted)
	}
}

func TestMapMappingRoutesRequireAuthentication(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/map-mappings", nil)
	recorder := httptest.NewRecorder()

	mapMappingTestApp(&mapMappingUseCaseStub{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
