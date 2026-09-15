package usecase

import (
	"context"
	"fmt"

	"hackz-mosa-back/internal/domain"
)

type SignIn struct {
	verifier domain.TokenVerifier
	users    domain.UserRepository
}

func NewSignIn(verifier domain.TokenVerifier, users domain.UserRepository) *SignIn {
	return &SignIn{verifier: verifier, users: users}
}

func (uc *SignIn) Execute(ctx context.Context, idToken string) (domain.User, error) {
	identity, err := uc.verifier.Verify(ctx, idToken)
	if err != nil {
		return domain.User{}, fmt.Errorf("verify Firebase ID token: %w", domain.ErrInvalidToken)
	}
	if identity.FirebaseUID == "" {
		return domain.User{}, domain.ErrInvalidToken
	}
	if !identity.Provider.Supported() {
		return domain.User{}, domain.ErrUnsupportedProvider
	}

	user, err := uc.users.Upsert(ctx, identity)
	if err != nil {
		return domain.User{}, fmt.Errorf("upsert authenticated user: %w", err)
	}
	return user, nil
}
