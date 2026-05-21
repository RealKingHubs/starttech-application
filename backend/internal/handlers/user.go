package handlers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/auth"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/cache"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/config"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/models"
	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/utils"
)

// UserHandler holds dependencies for user-related handlers.
type UserHandler struct {
	collection     *mongo.Collection
	todoCollection *mongo.Collection
	tokenSvc       *auth.TokenService
	cache          cache.Cache
	dbClient       *mongo.Client // Added for cache refreshing
	config         config.Config // Added for cache refreshing
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(collection *mongo.Collection, todoCollection *mongo.Collection, tokenSvc *auth.TokenService, cache cache.Cache, db *mongo.Client, cfg config.Config) *UserHandler {
	return &UserHandler{
		collection:     collection,
		todoCollection: todoCollection,
		tokenSvc:       tokenSvc,
		cache:          cache,
		dbClient:       db,
		config:         cfg,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with the provided details
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user body models.RegisterUserDTO true "User Registration Info"
// @Success      201  {object}  map[string]interface{} "{'message': 'User registered successfully'}"
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      409  {object}  map[string]string "Username is already taken"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var dto models.RegisterUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username is already taken
	count, err := h.collection.CountDocuments(context.Background(), bson.M{"username": strings.ToLower(dto.Username)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken"})
		return
	}

	now := time.Now()

	newUser := models.User{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Username:  strings.ToLower(dto.Username),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := newUser.HashPassword(dto.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	_, err = h.collection.InsertOne(context.Background(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Username is now taken, so cache this information
	usernameCacheKey := fmt.Sprintf("username-taken:%s", newUser.Username)
	h.cache.Set(context.Background(), usernameCacheKey, true, 5*time.Minute)

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login godoc
// @Summary      Log in a user
// @Description  Logs in a user with username and password, returning a session token.
// @Description  The token is returned in the response body and as an httpOnly cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body models.LoginUserDTO true "User Login Credentials"
// @Success      200  {object} map[string]interface{} "Returns a success message, the JWT token, and user details"
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      401  {object}  map[string]string "Invalid username or password"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var dto models.LoginUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.collection.FindOne(context.Background(), bson.M{"username": strings.ToLower(dto.Username)}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if !user.CheckPasswordHash(dto.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := h.tokenSvc.GenerateToken(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set httpOnly cookie for web clients
	cookieDomain := utils.GetCookieDomain(c, h.config.CookieDomains)
	c.SetCookie("token", token, h.tokenSvc.GetExpirationSeconds(), "/", cookieDomain, h.config.SecureCookie, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token, // Also return token in body for API clients
		"user": models.PublicUser{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Username:  user.Username,
		},
	})
}

// Logout godoc
// @Summary      Log out a user
// @Description  Clears the user's session cookie
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]string "{'message': 'Logged out successfully'}"
// @Router       /auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	cookieDomain := utils.GetCookieDomain(c, h.config.CookieDomains)
	c.SetCookie("token", "", -1, "/", cookieDomain, h.config.SecureCookie, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// UpdateUser godoc
// @Summary      Update current user's profile
// @Description  Updates the first name, last name, and/or username of the authenticated user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        user body models.UpdateUserDTO true "User Update Info"
// @Success      200  {object}  map[string]string "{'message': 'Profile updated successfully'}"
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      409  {object}  map[string]string "Username is already taken"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /users/me [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDHex, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	var dto models.UpdateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.D{}

	// Handle username change
	if dto.Username != nil {
		newUsername := strings.ToLower(*dto.Username)
		if len(newUsername) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be at least 3 characters"})
			return
		}
		// Check if the new username is already taken by ANOTHER user
		filter := bson.M{
			"username": newUsername,
			"_id":      bson.M{"$ne": userID}, // Check for users other than the current one
		}
		count, err := h.collection.CountDocuments(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking username"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Username is already taken"})
			return
		}
		update = append(update, bson.E{Key: "username", Value: newUsername})
	}

	if dto.FirstName != nil {
		update = append(update, bson.E{Key: "firstName", Value: *dto.FirstName})
	}
	if dto.LastName != nil {
		update = append(update, bson.E{Key: "lastName", Value: *dto.LastName})
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No update fields provided"})
		return
	}

	update = append(update, bson.E{Key: "updatedAt", Value: primitive.NewDateTimeFromTime(time.Now())})

	filter := bson.M{"_id": userID}
	result, err := h.collection.UpdateOne(context.Background(), filter, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// ChangePassword godoc
// @Summary      Change current user's password
// @Description  Allows an authenticated user to change their password
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        passwords body models.ChangePasswordDTO true "Old and New Passwords"
// @Success      200  {object}  map[string]string "{'message': 'Password changed successfully'}"
// @Failure      400  {object}  map[string]string "Invalid input or validation error"
// @Failure      401  {object}  map[string]string "Unauthorized or incorrect old password"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDHex, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	var dto models.ChangePasswordDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if dto.OldPassword == dto.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password cannot be the same as the old password"})
		return
	}

	// Fetch the user from the database
	var user models.User
	filter := bson.M{"_id": userID}
	err = h.collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if the old password is correct
	if !user.CheckPasswordHash(dto.OldPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password"})
		return
	}

	// Hash the new password
	if err := user.HashPassword(dto.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update the password in the database
	update := bson.M{"$set": bson.M{"password": user.Password}}
	_, err = h.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// DeleteUser godoc
// @Summary      Delete current user's account
// @Description  Permanently deletes the authenticated user's account and all their associated data (e.g., todos)
// @Tags         users
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]string "{'message': 'Account deleted successfully'}"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /users/me [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDHex, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	// Use a transaction to ensure both the user and their todos are deleted
	session, err := h.dbClient.StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start database session"})
		return
	}
	defer session.EndSession(context.Background())

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Delete all todos for this user
		_, err := h.todoCollection.DeleteMany(sessCtx, bson.M{"userId": userID})
		if err != nil {
			return nil, err
		}

		// Delete the user
		result, err := h.collection.DeleteOne(sessCtx, bson.M{"_id": userID})
		if err != nil {
			return nil, err
		}

		if result.DeletedCount == 0 {
			return nil, mongo.ErrNoDocuments // Use a standard error to indicate user not found
		}

		return result, nil
	}

	_, err = session.WithTransaction(context.Background(), callback)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account: " + err.Error()})
		return
	}

	// Clear the session cookie
	cookieDomain := utils.GetCookieDomain(c, h.config.CookieDomains)
	c.SetCookie("token", "", -1, "/", cookieDomain, false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

// CheckUsernameAvailability godoc
// @Summary      Check if a username is available
// @Description  Checks the database and cache to see if a username is already in use
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        username path string true "Username to check"
// @Success      200  {object}  map[string]interface{} "Returns true if the username is available"
// @Failure      400  {object}  map[string]string "Username is too short"
// @Failure      500  {object}  map[string]string "Database error"
// @Router       /auth/username-check/{username} [get]
func (h *UserHandler) CheckUsernameAvailability(c *gin.Context) {
	// Trigger a random cache refresh in the background
	h.triggerRandomCacheRefresh()

	username := strings.ToLower(c.Param("username"))
	if len(username) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"available": false, "message": "Username must be at least 3 characters"})
		return
	}

	cacheKey := fmt.Sprintf("username-taken:%s", username)

	// 1. Check cache first
	var isTaken bool
	err := h.cache.Get(context.Background(), cacheKey, &isTaken)
	if err == nil && isTaken { // Cache hit and username is taken
		c.JSON(http.StatusOK, gin.H{"available": false, "message": "Username not available"})
		return
	}

	// 2. If not in cache, check database
	count, err := h.collection.CountDocuments(context.Background(), bson.M{"username": username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		// 3. Set cache for future requests
		h.cache.Set(context.Background(), cacheKey, true, 24*time.Hour)
		c.JSON(http.StatusOK, gin.H{"available": false, "message": "Username not available"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available": true, "message": "Username is available"})
}

// GetCurrentUser godoc
// @Summary      Get current user's profile
// @Description  Retrieves the profile of the user corresponding to the provided JWT
// @Tags         users
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  models.PublicUser
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      404  {object}  map[string]string "User not found"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userIDHex, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in token"})
		return
	}

	var user models.User
	err = h.collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, models.PublicUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	})
}

// triggerRandomCacheRefresh starts a background refresh of the username cache
// based on a random chance to keep the cache fresh over time.
func (h *UserHandler) triggerRandomCacheRefresh() {
	if !h.config.EnableCache {
		return
	}

	// Use a 5% chance to trigger a refresh (5 out of 100).
	if rand.Intn(100) < 5 {
		go func() {
			log.Println("Probabilistic cache refresh triggered...")
			ctx := context.Background()
			userCollection := h.dbClient.Database(h.config.DBName).Collection("users")

			opts := options.Find().SetProjection(bson.M{"username": 1})
			cursor, err := userCollection.Find(ctx, bson.D{}, opts)
			if err != nil {
				log.Printf("Error during cache refresh query: %v", err)
				return
			}
			defer cursor.Close(ctx)

			usernamesToCache := make(map[string]interface{})
			for cursor.Next(ctx) {
				var result struct {
					Username string `bson:"username"`
				}
				if err := cursor.Decode(&result); err == nil && result.Username != "" {
					cacheKey := fmt.Sprintf("username-taken:%s", result.Username)
					usernamesToCache[cacheKey] = true
				}
			}

			if len(usernamesToCache) > 0 {
				err := h.cache.SetMany(ctx, usernamesToCache, 24*time.Hour)
				if err == nil {
					h.cache.Set(ctx, "username_cache_initialized", "true", 24*time.Hour)
					log.Printf("Successfully refreshed %d usernames in cache.", len(usernamesToCache))
				}
			}
		}()
	}
}
