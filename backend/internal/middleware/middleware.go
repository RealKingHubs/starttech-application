package middleware

import (
	"net/http"
	"strings"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/auth"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/config"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates a gin.HandlerFunc for JWT authentication.
func AuthMiddleware(tokenSvc *auth.TokenService, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Try to get the token from the httpOnly cookie first
		cookie, err := c.Cookie("token")
		if err == nil && cookie != "" {
			tokenString = cookie
		} else {
			// 2. If no cookie, try to get from Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header not provided"})
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
				return
			}
			tokenString = parts[1]
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token found"})
			return
		}

		// 3. Validate the token
		userID, err := tokenSvc.ValidateToken(tokenString)
		if err != nil {
			// Clear invalid cookie if it exists
			cookieDomain := utils.GetCookieDomain(c, cfg.CookieDomains)
			c.SetCookie("token", "", -1, "/", cookieDomain, cfg.SecureCookie, true)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 4. Set user ID in the context for downstream handlers
		c.Set("userID", userID)

		c.Next()
	}
}
