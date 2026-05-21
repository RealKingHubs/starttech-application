package auth

import (
	"testing"
	// "time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenService provides unit tests for the JWT token service.
func TestTokenService(t *testing.T) {
	// Setup: Initialize the service with a secret key.
	secretKey := "test-secret-that-is-long-enough-for-hs256"
	expirationHours := 1
	tokenSvc := NewTokenService(secretKey, expirationHours)
	userID := "test-user-123"

	t.Run("Generate and Validate Token - Happy Path", func(t *testing.T) {
		// 1. Generate a token
		token, err := tokenSvc.GenerateToken(userID)
		require.NoError(t, err, "Token generation should not produce an error")
		require.NotEmpty(t, token, "Generated token should not be empty")

		// 2. Validate the token
		validatedUserID, err := tokenSvc.ValidateToken(token)
		require.NoError(t, err, "Token validation should not produce an error")
		require.NotNil(t, validatedUserID, "Validated UserID should not be nil")

		// 3. Assert claims are correct
		assert.Equal(t, userID, validatedUserID, "Validated UserID in claims should match the original UserID")
	})

	t.Run("Validate Token - Invalid Signature", func(t *testing.T) {
		// Create another service with a different key to simulate an invalid signature
		otherTokenSvc := NewTokenService("a-different-secret-key", expirationHours)
		token, err := otherTokenSvc.GenerateToken(userID)
		require.NoError(t, err)

		// Try to validate with the original service
		_, err = tokenSvc.ValidateToken(token)
		assert.Error(t, err, "Validation should fail for a token signed with a different key")
	})

	t.Run("Validate Token - Malformed Token", func(t *testing.T) {
		_, err := tokenSvc.ValidateToken("this.is.not.a.valid.token")
		assert.Error(t, err, "Validation should fail for a malformed token")
	})

	t.Run("Validate Token - No Bearer Prefix", func(t *testing.T) {
		// Generate a valid token but pass it without the "Bearer " prefix
		token, err := tokenSvc.GenerateToken(userID)
		require.NoError(t, err)

		_, err = tokenSvc.ValidateToken(token) // Assuming the service's ValidateToken expects the raw token
		assert.NoError(t, err, "Validation should succeed even without the Bearer prefix in this context")
	})
}
