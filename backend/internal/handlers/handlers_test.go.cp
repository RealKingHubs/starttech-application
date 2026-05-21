//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	// "time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/auth"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/cache"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/config"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/models"
)

var (
	testDBClient *mongo.Client
	testRedisURI string
	testRouter   *gin.Engine
	testCfg      config.Config
)

// TestMain sets up the test environment (Docker containers, DB connections)
// before any tests in this package are run, and tears it down afterwards.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start MongoDB container
	mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage("mongo:6.0"))
	if err != nil {
		log.Fatalf("failed to start mongodb container: %s", err)
	}
	mongoURI, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get mongodb connection string: %s", err)
	}

	// 2. Start Redis container
	redisContainer, err := redis.RunContainer(ctx, testcontainers.WithImage("redis:7"))
	if err != nil {
		log.Fatalf("failed to start redis container: %s", err)
	}
	testRedisURI, err = redisContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get redis connection string: %s", err)
	}

	// 3. Set up test configuration
	testCfg = config.Config{
		MongoURI:        mongoURI,
		DBName:       "test_db",
		EnableCache:  true,
		RedisAddr:    testRedisURI,
		JWTSecretKey: "a-secure-test-secret-key-for-integration-tests",
		JWTExpirationHours: 1,
	}

	// 4. Connect to the test database
	testDBClient, err = mongo.Connect(ctx, options.Client().ApplyURI(testCfg.MongoURI))
	if err != nil {
		log.Fatalf("failed to connect to test mongodb: %s", err)
	}

	// 5. Set up router for tests
	tokenSvc := auth.NewTokenService(testCfg.JWTSecretKey, testCfg.JWTExpirationHours)
	cacheSvc := cache.NewCacheService(testCfg)
	userCollection := testDBClient.Database(testCfg.DBName).Collection("users")
	todoCollection := testDBClient.Database(testCfg.DBName).Collection("todos")
	userHandler := NewUserHandler(userCollection, todoCollection, tokenSvc, cacheSvc, testDBClient, testCfg)

	gin.SetMode(gin.TestMode)
	testRouter = gin.New()
	authRoutes := testRouter.Group("/auth")
	{
		authRoutes.POST("/register", userHandler.Register)
	}

	// Run the tests
	exitCode := m.Run()

	// Teardown: Clean up resources
	if err := mongodbContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate mongodb container: %s", err)
	}
	if err := redisContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate redis container: %s", err)
	}

	os.Exit(exitCode)
}

// TestRegisterUser provides an integration test for the user registration endpoint.
func TestRegisterUser(t *testing.T) {
	// Teardown: ensure the collection is clean after the test.
	defer func() {
		err := testDBClient.Database(testCfg.DBName).Collection("users").Drop(context.Background())
		require.NoError(t, err)
	}()

	t.Run("Successful Registration", func(t *testing.T) {
		// 1. Prepare request body
		registerDTO := models.RegisterUserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Username:  "johndoe",
			Password:  "password123",
		}
		body, err := json.Marshal(registerDTO)
		require.NoError(t, err)

		// 2. Create the HTTP request
		req, err := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		require.NoError(t, err)

		// 3. Record the response
		rec := httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)

		// 4. Assertions
		assert.Equal(t, http.StatusCreated, rec.Code, "Expected status code 201 Created")

		var responseUser models.PublicUser
		err = json.Unmarshal(rec.Body.Bytes(), &responseUser)
		require.NoError(t, err)

		assert.Equal(t, registerDTO.FirstName, responseUser.FirstName)
		assert.Equal(t, registerDTO.Username, responseUser.Username)
		assert.NotEmpty(t, responseUser.ID)
	})

	t.Run("Registration with existing username", func(t *testing.T) {
		// 1. First, create a user successfully
		firstRegisterDTO := models.RegisterUserDTO{
			FirstName: "Jane",
			LastName:  "Doe",
			Username:  "janedoe",
			Password:  "password123",
		}
		body, err := json.Marshal(firstRegisterDTO)
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		// 2. Now, try to register again with the same username
		secondRegisterDTO := models.RegisterUserDTO{
			FirstName: "Janet",
			LastName:  "Smith",
			Username:  "janedoe", // Same username
			Password:  "newpassword456",
		}
		body, err = json.Marshal(secondRegisterDTO)
		require.NoError(t, err)
		req, err = http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		require.NoError(t, err)
		rec = httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)

		// 3. Assertions
		assert.Equal(t, http.StatusConflict, rec.Code, "Expected status code 409 Conflict")

		var errorResponse map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse["error"], "username is already taken")
	})
}
