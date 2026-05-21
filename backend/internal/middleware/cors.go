package middleware

import (
	"regexp"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var starttechS3WebsiteOriginPattern = regexp.MustCompile(`^http://[a-z0-9-]*starttech-frontend-[a-z0-9]+\.s3-website-[a-z0-9-]+\.amazonaws\.com$`)

func isExplicitlyAllowedOrigin(origin string, allowedOrigins []string) bool {
	normalizedOrigin := strings.TrimRight(strings.TrimSpace(origin), "/")

	for _, allowedOrigin := range allowedOrigins {
		normalizedAllowedOrigin := strings.TrimRight(strings.TrimSpace(allowedOrigin), "/")
		if normalizedAllowedOrigin != "" && normalizedAllowedOrigin == normalizedOrigin {
			return true
		}
	}

	return false
}

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
		AllowOriginFunc: func(origin string) bool {
			if isExplicitlyAllowedOrigin(origin, allowedOrigins) {
				return true
			}

			return starttechS3WebsiteOriginPattern.MatchString(origin)
		},
	}

	return cors.New(config)
}
