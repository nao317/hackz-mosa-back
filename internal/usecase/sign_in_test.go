package usecase

import (
	"context"
	"errors"
	"testing"

	"hackz-mosa-back/internal/domain"
)

type verifierStub struct {
	identity domain.Identity
	err      error
}

func (s verifierStub) Verify(context.Context, string) (domain.Identity, error) {
	return s.identity, s.err
}

type userRepositoryStub struct {
	identity domain.Identity
	called   bool
}

func (s *userRepositoryStub) Upsert(_ context.Context, identity domain.Identity) (domain.User, error) {
	s.called = true
	s.identity = identity
	return domain.User{FirebaseUID: identity.FirebaseUID, Provider: identity.Provider}, nil
}

func TestSignInSupportsConfiguredProviders(t *testing.T) {
	providers := []domain.AuthProvider{
		domain.ProviderEmail,
		domain.ProviderGoogle,
	}

	for _, provider := range providers {
		t.Run(string(provider), func(t *testing.T) {
			repository := &userRepositoryStub{}
			service := NewSignIn(verifierStub{identity: domain.Identity{
				FirebaseUID: "firebase-user",
				Provider:    provider,
			}}, repository)

			user, err := service.Execute(context.Background(), "valid-token")
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if !repository.called || user.Provider != provider {
				t.Fatalf("provider = %q, repository called = %v", user.Provider, repository.called)
			}
		})
	}
}

func TestSignInRejectsInvalidToken(t *testing.T) {
	repository := &userRepositoryStub{}
	service := NewSignIn(verifierStub{err: errors.New("invalid")}, repository)

	_, err := service.Execute(context.Background(), "invalid-token")

	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("error = %v, want ErrInvalidToken", err)
	}
	if repository.called {
		t.Fatal("repository must not be called")
	}
}

func TestSignInRejectsUnsupportedProvider(t *testing.T) {
	repository := &userRepositoryStub{}
	service := NewSignIn(verifierStub{identity: domain.Identity{
		FirebaseUID: "firebase-user",
		Provider:    "apple.com",
	}}, repository)

	_, err := service.Execute(context.Background(), "valid-token")

	if !errors.Is(err, domain.ErrUnsupportedProvider) {
		t.Fatalf("error = %v, want ErrUnsupportedProvider", err)
	}
	if repository.called {
		t.Fatal("repository must not be called")
	}
}
