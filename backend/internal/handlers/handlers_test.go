//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/auth"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/cache"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/config"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/database"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/logger"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/models"
)

// UserHandlerIntegrationTestSuite is the suite for integration tests.
type UserHandlerIntegrationTestSuite struct {
	suite.Suite
	db           *mongo.Client
	cacheService cache.Cache
	router       *gin.Engine
	cfg          config.Config
	// Containers
	mongoContainer *mongodb.MongoDBContainer
	redisContainer *redis.RedisContainer
}

// SetupSuite runs once before all tests in the suite.
func (s *UserHandlerIntegrationTestSuite) SetupSuite() {
	ctx := context.Background()

	// Initialize Logger
	logger.InitLogger(config.Config{LogLevel: "debug", LogFormat: "text"})
	slog.Info("Setting up integration test suite...")

	// Start MongoDB container
	mongoContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6.0"))
	s.Require().NoError(err, "Failed to start MongoDB container")
	s.mongoContainer = mongoContainer

	// Start Redis container
	redisContainer, err := redis.RunContainer(ctx, testcontainers.WithImage("redis:7"))
	s.Require().NoError(err, "Failed to start Redis container")
	s.redisContainer = redisContainer

	// Get the connection strings from the containers
	mongoURI, err := mongoContainer.ConnectionString(ctx)
	s.Require().NoError(err)
	redisFullURI, err := redisContainer.ConnectionString(ctx)
	s.Require().NoError(err)

	// --- FIX: Parse the full Redis URI to get just host:port ---
	redisURL, err := url.Parse(redisFullURI)
	s.Require().NoError(err, "Failed to parse Redis URI")
	redisAddr := redisURL.Host // This extracts "localhost:xxxxx"

	// Create a test config
	s.cfg = config.Config{
		MongoURI:         mongoURI,
		DBName:           "testdb",
		EnableCache:      true,
		RedisAddr:        redisAddr, // Use the corrected address
		JWTSecretKey:     "a-secure-test-secret-key-that-is-long",
		JWTExpirationHours: 1,
	}

	// Connect to the test database
	dbClient, err := database.ConnectMongo(s.cfg.MongoURI, s.cfg.DBName)
	s.Require().NoError(err, "Failed to connect to test MongoDB")
	s.db = dbClient

	// Initialize services
	s.cacheService = cache.NewCacheService(s.cfg)
	tokenService := auth.NewTokenService(s.cfg.JWTSecretKey, s.cfg.JWTExpirationHours)

	// Setup router
	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	userCollection := s.db.Database(s.cfg.DBName).Collection("users")
	todoCollection := s.db.Database(s.cfg.DBName).Collection("todos")

	userHandler := NewUserHandler(userCollection, todoCollection, tokenService, s.cacheService, s.db, s.cfg)

	// Setup routes for testing
	authRoutes := s.router.Group("/auth")
	{
		authRoutes.POST("/register", userHandler.Register)
	}
}

// TearDownSuite runs once after all tests in the suite have finished.
func (s *UserHandlerIntegrationTestSuite) TearDownSuite() {
	slog.Info("Tearing down integration test suite...")
	ctx := context.Background()
	s.Require().NoError(s.db.Disconnect(ctx))
	s.Require().NoError(s.mongoContainer.Terminate(ctx))
	s.Require().NoError(s.redisContainer.Terminate(ctx))
}

// TearDownTest runs after each test to clean up the database.
func (s *UserHandlerIntegrationTestSuite) TearDownTest() {
	ctx := context.Background()
	userCollection := s.db.Database(s.cfg.DBName).Collection("users")
	_, err := userCollection.DeleteMany(ctx, gin.H{})
	s.Require().NoError(err)
}

// TestUserHandlerIntegrationTestSuite runs the entire suite.
func TestUserHandlerIntegrationTestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("Skipping integration tests: set INTEGRATION environment variable to run")
	}
	suite.Run(t, new(UserHandlerIntegrationTestSuite))
}

// TestRegisterUser tests the user registration endpoint.
func (s *UserHandlerIntegrationTestSuite) TestRegisterUser_Success() {
	// Define the payload
	payload := models.RegisterUserDTO{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Password:  "password123",
	}
	body, _ := json.Marshal(payload)

	// Create the request
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Assert the response
	s.Equal(http.StatusCreated, w.Code, "Expected status code 201")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Equal("User registered successfully", response["message"])
	// s.NotNil(response["user"])

	// Verify the user was created in the database
	ctx := context.Background()
	userCollection := s.db.Database(s.cfg.DBName).Collection("users")
	var user models.User
	err = userCollection.FindOne(ctx, gin.H{"username": "johndoe"}).Decode(&user)
	s.Require().NoError(err, "User should exist in the database after registration")
	s.Equal("John", user.FirstName)
}
