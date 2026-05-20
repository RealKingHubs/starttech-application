package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {

	origins := os.Getenv("ALLOWED_ORIGINS")

	if origins == "" {
		origins = "http://localhost:5173"
	}

	allowedOrigins := strings.Split(origins, ",")

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

		MaxAge: 12 * time.Hour,
	}

	return cors.New(config)
}