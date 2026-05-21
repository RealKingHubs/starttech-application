package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCookieDomain determines the appropriate cookie domain based on the request host
// and the list of allowed domains.
func GetCookieDomain(c *gin.Context, allowedDomains []string) string {
	host := c.Request.Host
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	for _, domain := range allowedDomains {
		if domain == host {
			return domain
		}
		// Also allow subdomains if the allowed domain starts with a dot (optional, but good practice)
		// or if we want to be strict, just exact match.
		// For now, let's do exact match as per "list of allowed hosts".
	}

	// Fallback: if the host is not in the allowed list, we might default to the first allowed domain
	// or return empty string to let the browser decide (usually implies current host).
	// Given the user's request for "allowed hosts", returning the first one as a safe default
	// or the host itself if we want to be permissive for development might be okay.
	// But to be safe and strictly follow the config:
	if len(allowedDomains) > 0 {
		return allowedDomains[0]
	}

	return "localhost"
}
