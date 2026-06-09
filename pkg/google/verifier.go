package google

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/idtoken"
    "github.com/golang-jwt/jwt/v5"
)

type Verifier struct {
	allowedClientIDs []string
}

func NewVerifier(allowedClientIDs []string) *Verifier {
	return &Verifier{allowedClientIDs: allowedClientIDs}
}

func (v *Verifier) VerifyToken(ctx context.Context, idToken string) (string, string, error) {
	//Print aud claim
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err == nil {
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
			fmt.Printf("\n[DEBUG] Incoming Google Token 'aud' claim is: %v\n", claims["aud"])
			fmt.Printf("[DEBUG] Backend expected 'aud' list: %v\n\n", v.allowedClientIDs)
		}
	}
	
	payload, err := idtoken.Validate(ctx, idToken, "")
	if err != nil {
		return "", "", fmt.Errorf("token validation failed: %w", err)
	}

	isValidAudience := false
	for _, id := range v.allowedClientIDs {
		if payload.Audience == id {
			isValidAudience = true
			break
		}
	}

	if !isValidAudience {
		return "", "", fmt.Errorf("token validation failed: audience %s is not allowed", payload.Audience)
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return "", "", errors.New("email claim missing in Google token")
	}

	name, _ := payload.Claims["name"].(string)

	return email, name, nil
}