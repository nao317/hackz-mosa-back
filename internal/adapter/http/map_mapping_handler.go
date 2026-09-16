package http

import (
	"context"
	"errors"
	"log/slog"
	stdhttp "net/http"

	"github.com/labstack/echo/v5"
	"hackz-mosa-back/internal/domain"
)

type mapMappingUseCase interface {
	Create(context.Context, int64, domain.NewMapMapping) (domain.MapMapping, error)
	List(context.Context, int64) ([]domain.MapMapping, error)
	Delete(context.Context, int64, int64) error
}

type MapMappingHandler struct {
	authenticate signInUseCase
	mappings     mapMappingUseCase
}

func NewMapMappingHandler(authenticate signInUseCase, mappings mapMappingUseCase) *MapMappingHandler {
	return &MapMappingHandler{authenticate: authenticate, mappings: mappings}
}

func (h *MapMappingHandler) Register(group *echo.Group) {
	group.GET("/map-mappings", h.list)
	group.POST("/map-mappings", h.create)
	group.DELETE("/map-mappings/:mappingId", h.delete)
}

type mapMappingTrackRequest struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Artist          string  `json:"artist"`
	ArtworkURL      *string `json:"artworkUrl"`
	AudiusURL       *string `json:"audiusUrl"`
	DurationSeconds *int    `json:"durationSeconds"`
	StreamURL       string  `json:"streamUrl"`
	Source          string  `json:"source"`
}

type createMapMappingRequest struct {
	Area  []domain.GeoPoint      `json:"area"`
	Track mapMappingTrackRequest `json:"track"`
}

func (h *MapMappingHandler) create(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	var request createMapMappingRequest
	if err := c.Bind(&request); err != nil {
		return errorResponse(c, stdhttp.StatusBadRequest, "invalid_request", "request body must be valid JSON")
	}
	mapping, err := h.mappings.Create(c.Request().Context(), user.ID, domain.NewMapMapping{
		Area: request.Area,
		Track: domain.MappedTrack{
			TrackID:         request.Track.ID,
			Title:           request.Track.Title,
			Artist:          request.Track.Artist,
			ArtworkURL:      request.Track.ArtworkURL,
			AudiusURL:       request.Track.AudiusURL,
			DurationSeconds: request.Track.DurationSeconds,
			StreamURL:       request.Track.StreamURL,
			Source:          request.Track.Source,
		},
	})
	if err != nil {
		return mapMappingErrorResponse(c, err)
	}
	return c.JSON(stdhttp.StatusCreated, map[string]domain.MapMapping{"mapping": mapping})
}

func (h *MapMappingHandler) list(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	mappings, err := h.mappings.List(c.Request().Context(), user.ID)
	if err != nil {
		return mapMappingErrorResponse(c, err)
	}
	return c.JSON(stdhttp.StatusOK, map[string][]domain.MapMapping{"mappings": mappings})
}

func (h *MapMappingHandler) delete(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	mappingID, ok := pathID(c, "mappingId")
	if !ok {
		return errorResponse(c, stdhttp.StatusNotFound, "not_found", "map mapping was not found")
	}
	if err := h.mappings.Delete(c.Request().Context(), user.ID, mappingID); err != nil {
		return mapMappingErrorResponse(c, err)
	}
	return c.NoContent(stdhttp.StatusNoContent)
}

func (h *MapMappingHandler) currentUser(c *echo.Context) (domain.User, error) {
	idToken, ok := bearerToken(c.Request().Header.Get("Authorization"))
	if !ok {
		return domain.User{}, errorResponse(c, stdhttp.StatusUnauthorized, "unauthorized", "valid Firebase ID token is required")
	}
	user, err := h.authenticate.Execute(c.Request().Context(), idToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidToken):
			return domain.User{}, errorResponse(c, stdhttp.StatusUnauthorized, "unauthorized", "valid Firebase ID token is required")
		case errors.Is(err, domain.ErrUnsupportedProvider):
			return domain.User{}, errorResponse(c, stdhttp.StatusForbidden, "unsupported_provider", "authentication provider is not supported")
		default:
			slog.Error("map mapping authentication failed", "error", err)
			return domain.User{}, errorResponse(c, stdhttp.StatusInternalServerError, "internal_error", "internal server error")
		}
	}
	return user, nil
}

func mapMappingErrorResponse(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidMapMapping):
		return errorResponse(c, stdhttp.StatusUnprocessableEntity, "invalid_map_mapping", "area or track data is invalid")
	case errors.Is(err, domain.ErrMapMappingNotFound):
		return errorResponse(c, stdhttp.StatusNotFound, "not_found", "map mapping was not found")
	default:
		slog.Error("map mapping request failed", "error", err)
		return errorResponse(c, stdhttp.StatusInternalServerError, "internal_error", "internal server error")
	}
}
