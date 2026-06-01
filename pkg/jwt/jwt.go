package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct{
	secret string
}

func NewService(secret string) *Service {
	return &Service{secret: secret}
}

func (s *Service) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp": time.Now().Add(720 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(s.secret))
}

func (s *Service) ParseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", jwt.ErrInvalidKey
		}
		return userID, nil
	}

	return "", jwt.ErrInvalidKey
}
