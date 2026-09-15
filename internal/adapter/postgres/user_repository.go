package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"hackz-mosa-back/internal/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    firebase_uid TEXT NOT NULL UNIQUE,
    email TEXT,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    display_name TEXT,
    photo_url TEXT,
    auth_provider TEXT NOT NULL CHECK (auth_provider IN ('password', 'google.com', 'apple.com')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS users_email_idx ON users (email) WHERE email IS NOT NULL;
`

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(ctx context.Context, databaseURL string) (*UserRepository, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	return &UserRepository{pool: pool}, nil
}

func (r *UserRepository) Close() {
	r.pool.Close()
}

func (r *UserRepository) Migrate(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("migrate users schema: %w", err)
	}
	return nil
}

func (r *UserRepository) Upsert(ctx context.Context, identity domain.Identity) (domain.User, error) {
	const query = `
INSERT INTO users (
    firebase_uid, email, email_verified, display_name, photo_url, auth_provider
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (firebase_uid) DO UPDATE SET
    email = EXCLUDED.email,
    email_verified = EXCLUDED.email_verified,
    display_name = EXCLUDED.display_name,
    photo_url = EXCLUDED.photo_url,
    auth_provider = EXCLUDED.auth_provider,
    updated_at = NOW()
RETURNING id, firebase_uid, email, email_verified, display_name, photo_url,
          auth_provider, created_at, updated_at
`

	var (
		user                         domain.User
		email, displayName, photoURL pgtype.Text
	)
	err := r.pool.QueryRow(
		ctx,
		query,
		identity.FirebaseUID,
		nullable(identity.Email),
		identity.EmailVerified,
		nullable(identity.DisplayName),
		nullable(identity.PhotoURL),
		identity.Provider,
	).Scan(
		&user.ID,
		&user.FirebaseUID,
		&email,
		&user.EmailVerified,
		&displayName,
		&photoURL,
		&user.Provider,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}

	user.Email = textPointer(email)
	user.DisplayName = textPointer(displayName)
	user.PhotoURL = textPointer(photoURL)
	return user, nil
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
