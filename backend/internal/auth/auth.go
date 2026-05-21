package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService provides functionality for creating and validating JWTs.
type TokenService struct {
	secretKey     string
	expirationDur time.Duration
}

// NewTokenService creates a new instance of TokenService.
func NewTokenService(secret string, expirationHours int) *TokenService {
	return &TokenService{
		secretKey:     secret,
		expirationDur: time.Hour * time.Duration(expirationHours),
	}
}

// GenerateToken creates a new JWT for a given user ID.
func (s *TokenService) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID, // Subject (user ID)
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.expirationDur).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// ValidateToken parses and validates a token string.
// It returns the user ID (subject) if the token is valid.
func (s *TokenService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return "", errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", errors.New("invalid token claims: subject not found")
		}
		return userID, nil
	}

	return "", errors.New("invalid token")
}

// GetExpirationSeconds returns the token expiration duration in seconds.
func (s *TokenService) GetExpirationSeconds() int {
	return int(s.expirationDur.Seconds())
}
