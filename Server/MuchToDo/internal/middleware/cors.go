package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"strings"
)

// CORSMiddleware returns a CORS middleware using gin-contrib/cors
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = allowedOrigins
	config.AllowCredentials = true
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}

	return cors.New(config)
}

func CORSMiddleware2() gin.HandlerFunc {
	// Define allowed origins as a comma-separated string
	originsString := "http://localhost:5173,http://dev-starttech-frontend-ee4128bc.s3-website-us-east-1.amazonaws.com"
	var allowedOrigins []string
	if originsString != "" {
		// Split the originsString into individual origins and store them in allowedOrigins slice
		allowedOrigins = strings.Split(originsString, ",")
	}

	// Return the actual middleware handler function
	return func(c *gin.Context) {
		// Function to check if a given origin is allowed
		isOriginAllowed := func(origin string, allowedOrigins []string) bool {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			return false
		}

		// Get the Origin header from the request
		origin := c.Request.Header.Get("Origin")

		// Check if the origin is allowed
		if isOriginAllowed(origin, allowedOrigins) {
			// If the origin is allowed, set CORS headers in the response
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		}

		// Handle preflight OPTIONS requests by aborting with status 204
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Call the next handler
		c.Next()
	}
}
