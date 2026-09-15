package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"hackz-mosa-back/internal/domain"
)

func TestPlaylistRepositoryLifecycleAndTrackLimit(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	users, err := NewUserRepository(ctx, databaseURL)
	if err != nil {
		t.Fatalf("NewUserRepository() error = %v", err)
	}
	t.Cleanup(users.Close)
	if err := users.Migrate(ctx); err != nil {
		t.Fatalf("users.Migrate() error = %v", err)
	}
	playlists := NewPlaylistRepository(users)
	if err := playlists.Migrate(ctx); err != nil {
		t.Fatalf("playlists.Migrate() error = %v", err)
	}

	user, err := users.Upsert(ctx, domain.Identity{
		FirebaseUID: "playlist-integration-user",
		Provider:    domain.ProviderEmail,
	})
	if err != nil {
		t.Fatalf("users.Upsert() error = %v", err)
	}
	t.Cleanup(func() {
		if _, err := users.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, user.ID); err != nil {
			t.Errorf("clean up playlist user: %v", err)
		}
	})

	playlist, err := playlists.Create(ctx, user.ID, "Integration playlist", "repository lifecycle")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	for index := 1; index <= domain.MaxPlaylistTracks; index++ {
		playlist, err = playlists.AddTrack(ctx, user.ID, playlist.ID, integrationTrack(index))
		if err != nil {
			t.Fatalf("AddTrack(%d) error = %v", index, err)
		}
	}
	if playlist.TrackCount != domain.MaxPlaylistTracks {
		t.Fatalf("track count = %d, want %d", playlist.TrackCount, domain.MaxPlaylistTracks)
	}
	if _, err := playlists.AddTrack(ctx, user.ID, playlist.ID, integrationTrack(31)); !errors.Is(err, domain.ErrPlaylistFull) {
		t.Fatalf("31st AddTrack() error = %v, want ErrPlaylistFull", err)
	}

	reversed := make([]int64, len(playlist.Tracks))
	for index, track := range playlist.Tracks {
		reversed[len(reversed)-1-index] = track.ID
	}
	playlist, err = playlists.ReorderTracks(ctx, user.ID, playlist.ID, reversed)
	if err != nil {
		t.Fatalf("ReorderTracks() error = %v", err)
	}
	if playlist.Tracks[0].ID != reversed[0] || playlist.Tracks[0].Position != 1 {
		t.Fatalf("first track after reorder = %#v", playlist.Tracks[0])
	}

	deletedID := playlist.Tracks[10].ID
	playlist, err = playlists.DeleteTrack(ctx, user.ID, playlist.ID, deletedID)
	if err != nil {
		t.Fatalf("DeleteTrack() error = %v", err)
	}
	if playlist.TrackCount != domain.MaxPlaylistTracks-1 {
		t.Fatalf("track count after delete = %d", playlist.TrackCount)
	}
	for index, track := range playlist.Tracks {
		if track.ID == deletedID || track.Position != index+1 {
			t.Fatalf("track %d after delete = %#v", index, track)
		}
	}

	listed, err := playlists.List(ctx, user.ID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 1 || listed[0].TrackCount != domain.MaxPlaylistTracks-1 {
		t.Fatalf("listed playlists = %#v", listed)
	}
	if err := playlists.Delete(ctx, user.ID, playlist.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := playlists.Get(ctx, user.ID, playlist.ID); !errors.Is(err, domain.ErrPlaylistNotFound) {
		t.Fatalf("Get() after delete error = %v, want ErrPlaylistNotFound", err)
	}
}

func integrationTrack(index int) domain.NewPlaylistTrack {
	return domain.NewPlaylistTrack{
		TrackID:   fmt.Sprintf("track-%02d", index),
		Title:     fmt.Sprintf("Track %02d", index),
		Artist:    "Integration Artist",
		StreamURL: fmt.Sprintf("https://api.audius.co/v1/tracks/track-%02d/stream", index),
		Source:    "audius",
	}
}
