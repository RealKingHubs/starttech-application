package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/cache"
)

// HealthHandler holds dependencies for health checks.
type HealthHandler struct {
	dbClient       *mongo.Client
	cacheClient    cache.Cache
	isCacheEnabled bool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *mongo.Client, cache cache.Cache, cacheEnabled bool) *HealthHandler {
	return &HealthHandler{
		dbClient:       db,
		cacheClient:    cache,
		isCacheEnabled: cacheEnabled,
	}
}

// CheckHealth godoc
// @Summary      Show the status of server connections
// @Description  get the status of the database and cache (if enabled)
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string "A map showing the status of each service"
// @Failure      503  {object}  map[string]string "A map showing the status of each service, one or more will be 'down'"
// @Router       /health [get]
func (h *HealthHandler) CheckHealth(c *gin.Context) {
	status := gin.H{
		"database": "down",
		"cache":    "disabled",
	}
	isHealthy := true

	// --- Check Database ---
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.dbClient.Ping(ctx, nil); err == nil {
		status["database"] = "ok"
	} else {
		isHealthy = false
	}

	// --- Check Cache (if enabled) ---
	if h.isCacheEnabled {
		status["cache"] = "down" // Assume down until proven otherwise
		ctxCache, cancelCache := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelCache()

		if err := h.cacheClient.Ping(ctxCache); err == nil {
			status["cache"] = "ok"
		} else {
			isHealthy = false
		}
	}

	if !isHealthy {
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}
