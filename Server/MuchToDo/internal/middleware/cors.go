package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS configuration
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {

	// Fallback origins if config is empty
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{
			"http://localhost:5173",
			"http://dev-starttech-frontend-ee4128bc.s3-website-us-east-1.amazonaws.com",
		}
	}

	// Clean whitespace
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
	}

	config := cors.Config{
		AllowOrigins: allowedOrigins,

		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},

		ExposeHeaders: []string{
			"Content-Length",
		},

		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}

	return cors.New(config)
}