package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidToken        = errors.New("invalid authentication token")
	ErrUnsupportedProvider = errors.New("unsupported authentication provider")
)

type AuthProvider string

const (
	ProviderEmail  AuthProvider = "password"
	ProviderGoogle AuthProvider = "google.com"
	ProviderApple  AuthProvider = "apple.com"
)

func (p AuthProvider) Supported() bool {
	switch p {
	case ProviderEmail, ProviderGoogle, ProviderApple:
		return true
	default:
		return false
	}
}

type Identity struct {
	FirebaseUID   string
	Email         string
	EmailVerified bool
	DisplayName   string
	PhotoURL      string
	Provider      AuthProvider
}

type User struct {
	ID            int64        `json:"id"`
	FirebaseUID   string       `json:"firebaseUid"`
	Email         *string      `json:"email"`
	EmailVerified bool         `json:"emailVerified"`
	DisplayName   *string      `json:"displayName"`
	PhotoURL      *string      `json:"photoUrl"`
	Provider      AuthProvider `json:"provider"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

type TokenVerifier interface {
	Verify(ctx context.Context, idToken string) (Identity, error)
}

type UserRepository interface {
	Upsert(ctx context.Context, identity Identity) (User, error)
}
