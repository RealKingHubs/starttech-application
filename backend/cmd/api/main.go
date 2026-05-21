// @title           MuchToDo API
// @version         1.0
// @description     This is an API for MuchToDo application with user authentication.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support - Innocent
// @contact.url    https://github.com/Innocent9712
// @contact.email  innocent@altschoolafrica.com
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath  /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description "Type 'Bearer' followed by a space and a JWT token."
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/auth"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/cache"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/config"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/database"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/handlers"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/logger"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/middleware"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/routes"

	// Swagger imports
	_ "github.com/Innocent9712/much-to-do/Server/MuchToDo/docs" // This is required for swag to find your docs
)

const usernameCacheSentinelKey = "username_cache_initialized"
const usernameCacheTTL = 24 * time.Hour

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// --- Logger ---
	// This must be initialized before any other component that might log.
	logger.InitLogger(cfg)
	slog.Info("Logger initialized", "level", cfg.LogLevel, "format", cfg.LogFormat)

	// 2. Connect to Database
	dbClient, err := database.ConnectMongo(cfg.MongoURI, cfg.DBName)
	if err != nil {
		slog.Error("could not connect to MongoDB", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err = dbClient.Disconnect(context.Background()); err != nil {
			slog.Error("Error disconnecting from MongoDB", slog.Any("error", err))
		}
	}()
	slog.Info("Successfully connected to MongoDB.")

	// 3. Initialize Services (Cache, Auth)
	cacheService := cache.NewCacheService(cfg)
	tokenService := auth.NewTokenService(cfg.JWTSecretKey, cfg.JWTExpirationHours)

	// Preload usernames into cache if enabled
	preloadUsernamesIntoCache(dbClient, cacheService, cfg)

	// 4. Set up API router
	router := setupRouter(dbClient, cfg, tokenService, cacheService)

	// 5. Start Server with graceful shutdown
	startServer(router, cfg.ServerPort)
}

// preloadUsernamesIntoCache queries for all usernames and loads them into the cache,
// but only if caching is enabled and a sentinel key indicates the cache is empty.
func preloadUsernamesIntoCache(db *mongo.Client, cacheSvc cache.Cache, cfg config.Config) {
	if !cfg.EnableCache {
		slog.Info("Caching is disabled. Skipping username preloading.")
		return
	}

	// Check if the cache has already been initialized in this cycle.
	var sentinelVal string
	err := cacheSvc.Get(context.Background(), usernameCacheSentinelKey, &sentinelVal)
	if err == nil {
		slog.Info("Username cache already initialized. Skipping preload.")
		return
	}

	slog.Info("Preloading usernames into cache...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userCollection := db.Database(cfg.DBName).Collection("users")

	// Find all users, but only project the username field for efficiency
	opts := options.Find().SetProjection(bson.M{"username": 1})
	cursor, err := userCollection.Find(ctx, bson.D{}, opts)
	if err != nil {
		slog.Error("Error querying for usernames to preload", slog.Any("error", err))
		return
	}
	defer cursor.Close(ctx)

	// Use a map to prepare for batch cache insertion
	usernamesToCache := make(map[string]interface{})
	for cursor.Next(ctx) {
		var result struct {
			Username string `bson:"username"`
		}
		if err := cursor.Decode(&result); err != nil {
			slog.Warn("Error decoding username during preload", slog.Any("error", err))
			continue
		}
		if result.Username != "" {
			cacheKey := fmt.Sprintf("username-taken:%s", result.Username)
			usernamesToCache[cacheKey] = true
		}
	}

	if err := cursor.Err(); err != nil {
		slog.Error("Cursor error during username preload", slog.Any("error", err))
		return
	}

	if len(usernamesToCache) > 0 {
		err := cacheSvc.SetMany(ctx, usernamesToCache, usernameCacheTTL)
		if err != nil {
			slog.Error("Error preloading usernames to cache", slog.Any("error", err))
		} else {
			// Set the sentinel key to prevent re-loading until it expires.
			cacheSvc.Set(ctx, usernameCacheSentinelKey, "true", usernameCacheTTL)
			slog.Info("Successfully preloaded usernames into cache", "count", len(usernamesToCache))
		}
	} else {
		slog.Info("No usernames found to preload.")
	}
}

// setupRouter initializes the Gin router and sets up the routes.
func setupRouter(db *mongo.Client, cfg config.Config, tokenSvc *auth.TokenService, cacheSvc cache.Cache) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Initialize collections
	todoCollection := db.Database(cfg.DBName).Collection("todos")
	userCollection := db.Database(cfg.DBName).Collection("users")

	// Initialize handlers
	todoHandler := handlers.NewTodoHandler(todoCollection)
	userHandler := handlers.NewUserHandler(userCollection, todoCollection, tokenSvc, cacheSvc, db, cfg)
	healthHandler := handlers.NewHealthHandler(db, cacheSvc, cfg.EnableCache)

	// Middleware
	corsMiddleware := middleware.CORSMiddleware(cfg.AllowedOrigins)
	// corsMiddleware := middleware.CORSMiddleware2()
	authMiddleware := middleware.AuthMiddleware(tokenSvc, cfg)

	// Apply CORS middleware to the router
	router.Use(corsMiddleware)

	// Register all routes
	routes.RegisterRoutes(router, userHandler, todoHandler, healthHandler, authMiddleware)

	// A simple ping route for health checks
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to MuchToDo API"})
	})

	// Test route to debug /todos issue
	router.GET("/test-todos", func(c *gin.Context) {
		println("=== TEST TODOS ROUTE HIT ===")
		c.JSON(http.StatusOK, gin.H{"message": "Test todos route works!"})
	})

	// Handle 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
	})

	return router
}

// startServer starts the HTTP server and handles graceful shutdown.
func startServer(router *gin.Engine, port string) {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		// Service connections
		slog.Info("Server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server listen error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Server exiting.")
}
