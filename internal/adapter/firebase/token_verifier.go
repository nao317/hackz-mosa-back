package firebase

import (
	"context"
	"fmt"

	firebaseadmin "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"hackz-mosa-back/internal/domain"
)

type TokenVerifier struct {
	client *auth.Client
}

func NewTokenVerifier(ctx context.Context, projectID string) (*TokenVerifier, error) {
	app, err := firebaseadmin.NewApp(ctx, &firebaseadmin.Config{ProjectID: projectID})
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase Auth client: %w", err)
	}
	return &TokenVerifier{client: client}, nil
}

func (v *TokenVerifier) Verify(ctx context.Context, idToken string) (domain.Identity, error) {
	token, err := v.client.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	if err != nil {
		return domain.Identity{}, fmt.Errorf("verify token: %w", err)
	}

	return domain.Identity{
		FirebaseUID:   token.UID,
		Email:         stringClaim(token.Claims, "email"),
		EmailVerified: boolClaim(token.Claims, "email_verified"),
		DisplayName:   stringClaim(token.Claims, "name"),
		PhotoURL:      stringClaim(token.Claims, "picture"),
		Provider:      domain.AuthProvider(token.Firebase.SignInProvider),
	}, nil
}

func stringClaim(claims map[string]interface{}, key string) string {
	value, _ := claims[key].(string)
	return value
}

func boolClaim(claims map[string]interface{}, key string) bool {
	value, _ := claims[key].(bool)
	return value
}
