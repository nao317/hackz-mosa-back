package postgres

import (
	"context"
	"os"
	"testing"

	"hackz-mosa-back/internal/domain"
)

func TestUserRepositoryUpsert(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	repository, err := NewUserRepository(ctx, databaseURL)
	if err != nil {
		t.Fatalf("NewUserRepository() error = %v", err)
	}
	t.Cleanup(repository.Close)

	if err := repository.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	t.Cleanup(func() {
		if _, err := repository.pool.Exec(ctx, "DELETE FROM users WHERE firebase_uid = $1", "integration-firebase-user"); err != nil {
			t.Errorf("clean up test user: %v", err)
		}
	})

	identity := domain.Identity{
		FirebaseUID:   "integration-firebase-user",
		Email:         "first@example.com",
		EmailVerified: true,
		DisplayName:   "First Name",
		Provider:      domain.ProviderGoogle,
	}
	first, err := repository.Upsert(ctx, identity)
	if err != nil {
		t.Fatalf("first Upsert() error = %v", err)
	}

	identity.Email = "updated@example.com"
	identity.Provider = domain.ProviderApple
	updated, err := repository.Upsert(ctx, identity)
	if err != nil {
		t.Fatalf("second Upsert() error = %v", err)
	}

	if updated.ID != first.ID {
		t.Fatalf("updated ID = %d, want %d", updated.ID, first.ID)
	}
	if updated.Email == nil || *updated.Email != identity.Email {
		t.Fatalf("updated email = %v, want %q", updated.Email, identity.Email)
	}
	if updated.Provider != domain.ProviderApple {
		t.Fatalf("updated provider = %q, want %q", updated.Provider, domain.ProviderApple)
	}
}
