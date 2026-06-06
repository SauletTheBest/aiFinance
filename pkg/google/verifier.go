package google

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/idtoken"
    "github.com/golang-jwt/jwt/v5"
)

type Verifier struct {
	clientID string
}

func NewVerifier(clientID string) *Verifier {
	return &Verifier{clientID: clientID}
}

func (v *Verifier) VerifyToken(ctx context.Context, idToken string) (string, string, error) {
	//Print aud claim
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err == nil {
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
			fmt.Printf("\n[DEBUG] Incoming Google Token 'aud' claim is: %v\n", claims["aud"])
			fmt.Printf("[DEBUG] Backend expected 'aud' claim is: %s\n\n", v.clientID)
		}
	}
	//
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