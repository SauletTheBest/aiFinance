package google

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/idtoken"
)

type Verifier struct {
	clientID string
}

func NewVerifier(clientID string) *Verifier {
	return &Verifier{clientID: clientID}
}

func (v *Verifier) VerifyToken(ctx context.Context, idToken string) (string, string, error) {
	payload, err := idtoken.Validate(ctx, idToken, v.clientID)
	if err != nil {
		return "", "", fmt.Errorf("token validation failed: %w", err)
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return "", "", errors.New("email claim missing in Google token")
	}

	name, _ := payload.Claims["name"].(string)


	return email, name, nil
}