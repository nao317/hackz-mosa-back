package http

import (
	"context"
	"errors"
	"log/slog"
	stdhttp "net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"hackz-mosa-back/internal/domain"
)

type playlistUseCase interface {
	Create(context.Context, int64, string, string) (domain.Playlist, error)
	List(context.Context, int64) ([]domain.Playlist, error)
	Get(context.Context, int64, int64) (domain.Playlist, error)
	Delete(context.Context, int64, int64) error
	AddTrack(context.Context, int64, int64, domain.NewPlaylistTrack) (domain.Playlist, error)
	ReorderTracks(context.Context, int64, int64, []int64) (domain.Playlist, error)
	DeleteTrack(context.Context, int64, int64, int64) (domain.Playlist, error)
}

type PlaylistHandler struct {
	authenticate signInUseCase
	playlists    playlistUseCase
}

func NewPlaylistHandler(authenticate signInUseCase, playlists playlistUseCase) *PlaylistHandler {
	return &PlaylistHandler{authenticate: authenticate, playlists: playlists}
}

func (h *PlaylistHandler) Register(group *echo.Group) {
	group.GET("/playlists", h.list)
	group.POST("/playlists", h.create)
	group.GET("/playlists/:playlistId", h.get)
	group.DELETE("/playlists/:playlistId", h.delete)
	group.POST("/playlists/:playlistId/tracks", h.addTrack)
	group.PUT("/playlists/:playlistId/tracks/order", h.reorderTracks)
	group.DELETE("/playlists/:playlistId/tracks/:itemId", h.deleteTrack)
}

type createPlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *PlaylistHandler) create(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	var request createPlaylistRequest
	if err := c.Bind(&request); err != nil {
		return errorResponse(c, stdhttp.StatusBadRequest, "invalid_request", "request body must be valid JSON")
	}
	playlist, err := h.playlists.Create(c.Request().Context(), user.ID, request.Name, request.Description)
	if err != nil {
		return playlistErrorResponse(c, err)
	}
	return c.JSON(stdhttp.StatusCreated, map[string]domain.Playlist{"playlist": playlist})
}

func (h *PlaylistHandler) list(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	playlists, err := h.playlists.List(c.Request().Context(), user.ID)
	if err != nil {
		return playlistErrorResponse(c, err)
	}
	return c.JSON(stdhttp.StatusOK, map[string][]domain.Playlist{"playlists": playlists})
}

func (h *PlaylistHandler) get(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	playlistID, ok := pathID(c, "playlistId")
	if !ok {
		return errorResponse(c, stdhttp.StatusNotFound, "not_found", "playlist was not found")
	}
	playlist, err := h.playlists.Get(c.Request().Context(), user.ID, playlistID)
	if err != nil {
		return playlistErrorResponse(c, err)
	}
	return c.JSON(stdhttp.StatusOK, map[string]domain.Playlist{"playlist": playlist})
}

func (h *PlaylistHandler) delete(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	playlistID, ok := pathID(c, "playlistId")
	if !ok {
		return errorResponse(c, stdhttp.StatusNotFound, "not_found", "playlist was not found")
	}
	if err := h.playlists.Delete(c.Request().Context(), user.ID, playlistID); err != nil {
		return playlistErrorResponse(c, err)
	}
	return c.NoContent(stdhttp.StatusNoContent)
}

type addTrackRequest struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Artist          string  `json:"artist"`
	ArtworkURL      *string `json:"artworkUrl"`
	AudiusURL       *string `json:"audiusUrl"`
	DurationSeconds *int    `json:"durationSeconds"`
	StreamURL       string  `json:"streamUrl"`
	Source          string  `json:"source"`
}

func (h *PlaylistHandler) addTrack(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	playlistID, ok := pathID(c, "playlistId")
	if !ok {
		return errorResponse(c, stdhttp.StatusNotFound, "not_found", "playlist was not found")
	}
	var request addTrackRequest
	if err := c.Bind(&request); err != nil {
		return errorResponse(c, stdhttp.StatusBadRequest, "invalid_request", "request body must be valid JSON")
	}
	playlist, err := h.playlists.AddTrack(c.Request().Context(), user.ID, playlistID, domain.NewPlaylistTrack{
		TrackID:         request.ID,
		Title:           request.Title,
		Artist:          request.Artist,
		ArtworkURL:      request.ArtworkURL,
		AudiusURL:       request.AudiusURL,
		DurationSeconds: request.DurationSeconds,
		StreamURL:       request.StreamURL,
		Source:          request.Source,
	})
	if err != nil {
		return playlistErrorResponse(c, err)
	}
	return c.JSON(stdhttp.StatusCreated, map[string]domain.Playlist{"playlist": playlist})
}

type reorderTracksRequest struct {
	ItemIDs []int64 `json:"itemIds"`
}

func (h *PlaylistHandler) reorderTracks(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	playlistID, ok := pathID(c, "playlistId")
	if !ok {
		return errorResponse(c, stdhttp.StatusNotFound, "not_found", "playlist was not found")
	}
	var request reorderTracksRequest
	if err := c.Bind(&request); err != nil {
		return errorResponse(c, stdhttp.StatusBadRequest, "invalid_request", "request body must be valid JSON")
	}
	playlist, err := h.playlists.ReorderTracks(c.Request().Context(), user.ID, playlistID, request.ItemIDs)
	if err != nil {
		return playlistErrorResponse(c, err)
	}
	return c.JSON(stdhttp.StatusOK, map[string]domain.Playlist{"playlist": playlist})
}

func (h *PlaylistHandler) deleteTrack(c *echo.Context) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	playlistID, playlistOK := pathID(c, "playlistId")
	itemID, itemOK := pathID(c, "itemId")
	if !playlistOK || !itemOK {
		return errorResponse(c, stdhttp.StatusNotFound, "not_found", "playlist or track was not found")
	}
	playlist, err := h.playlists.DeleteTrack(c.Request().Context(), user.ID, playlistID, itemID)
	if err != nil {
		return playlistErrorResponse(c, err)
	}
	return c.JSON(stdhttp.StatusOK, map[string]domain.Playlist{"playlist": playlist})
}

func (h *PlaylistHandler) currentUser(c *echo.Context) (domain.User, error) {
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
			slog.Error("playlist authentication failed", "error", err)
			return domain.User{}, errorResponse(c, stdhttp.StatusInternalServerError, "internal_error", "internal server error")
		}
	}
	return user, nil
}

func pathID(c *echo.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	return id, err == nil && id > 0
}

func playlistErrorResponse(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidPlaylist):
		return errorResponse(c, stdhttp.StatusUnprocessableEntity, "invalid_playlist", "playlist or track data is invalid")
	case errors.Is(err, domain.ErrPlaylistNotFound):
		return errorResponse(c, stdhttp.StatusNotFound, "not_found", "playlist or track was not found")
	case errors.Is(err, domain.ErrDuplicateTrack):
		return errorResponse(c, stdhttp.StatusConflict, "duplicate_track", "track is already in playlist")
	case errors.Is(err, domain.ErrPlaylistFull):
		return errorResponse(c, stdhttp.StatusUnprocessableEntity, "playlist_full", "a playlist can contain at most 30 tracks")
	case errors.Is(err, domain.ErrInvalidTrackOrder):
		return errorResponse(c, stdhttp.StatusUnprocessableEntity, "invalid_track_order", "itemIds must contain every playlist track exactly once")
	default:
		slog.Error("playlist request failed", "error", err)
		return errorResponse(c, stdhttp.StatusInternalServerError, "internal_error", "internal server error")
	}
}
