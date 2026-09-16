package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"hackz-mosa-back/internal/domain"
)

func TestMapMappingRepositoryLifecycle(t *testing.T) {
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
	repository := NewMapMappingRepository(users)
	if err := repository.Migrate(ctx); err != nil {
		t.Fatalf("repository.Migrate() error = %v", err)
	}

	user, err := users.Upsert(ctx, domain.Identity{
		FirebaseUID: "map-mapping-integration-user",
		Provider:    domain.ProviderEmail,
	})
	if err != nil {
		t.Fatalf("users.Upsert() error = %v", err)
	}
	t.Cleanup(func() {
		if _, err := users.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, user.ID); err != nil {
			t.Errorf("clean up map mapping user: %v", err)
		}
	})

	created, err := repository.Create(ctx, user.ID, domain.NewMapMapping{
		Area: []domain.GeoPoint{
			{Latitude: 35.68, Longitude: 139.76},
			{Latitude: 35.69, Longitude: 139.76},
			{Latitude: 35.69, Longitude: 139.77},
		},
		Track: domain.MappedTrack{
			TrackID:   "mapped-track",
			Title:     "Mapped song",
			Artist:    "Artist",
			StreamURL: "https://api.audius.co/v1/tracks/mapped-track/stream",
			Source:    "audius",
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	listed, err := repository.List(ctx, user.ID)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID || len(listed[0].Area) != 3 {
		t.Fatalf("listed mappings = %#v", listed)
	}
	if err := repository.Delete(ctx, user.ID, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := repository.Delete(ctx, user.ID, created.ID); !errors.Is(err, domain.ErrMapMappingNotFound) {
		t.Fatalf("second Delete() error = %v, want ErrMapMappingNotFound", err)
	}
}
