package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// =========================================
// STANDARD CORS MIDDLEWARE
// =========================================

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	config := cors.DefaultConfig()

	config.AllowOrigins = allowedOrigins
	config.AllowCredentials = true

	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
	}

	config.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"OPTIONS",
	}

	return cors.New(config)
}

// =========================================
// CUSTOM DYNAMIC CORS MIDDLEWARE
// =========================================

func CORSMiddleware2() gin.HandlerFunc {

	// Read allowed origins from environment variable
	originsString := os.Getenv("ALLOWED_ORIGINS")

	// Fallback for local development
	if originsString == "" {
		originsString = "http://localhost:5173"
	}

	var allowedOrigins []string

	if originsString != "" {
		allowedOrigins = strings.Split(originsString, ",")
	}

	return func(c *gin.Context) {

		origin := c.Request.Header.Get("Origin")

		isOriginAllowed := false

		for _, allowedOrigin := range allowedOrigins {

			if strings.TrimSpace(allowedOrigin) == origin {
				isOriginAllowed = true
				break
			}
		}

		if isOriginAllowed {

			c.Writer.Header().Set(
				"Access-Control-Allow-Origin",
				origin,
			)

			c.Writer.Header().Set(
				"Access-Control-Allow-Credentials",
				"true",
			)

			c.Writer.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
			)

			c.Writer.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, PATCH, DELETE, OPTIONS",
			)
		}

		// Handle preflight request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}