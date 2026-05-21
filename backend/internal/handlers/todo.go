package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Innocent9712/much-to-do/Server/MuchToDo/internal/models"
)

// TodoHandler holds the database collection for todos.
type TodoHandler struct {
	collection *mongo.Collection
}

// NewTodoHandler creates a new handler for ToDo operations.
func NewTodoHandler(collection *mongo.Collection) *TodoHandler {
	return &TodoHandler{collection: collection}
}

// getUserIDFromContext retrieves the user ID from the Gin context.
func getUserIDFromContext(c *gin.Context) (primitive.ObjectID, error) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID, errors.New("user ID not found in context")
	}

	userIDHex, ok := userIDVal.(string)
	if !ok {
		return primitive.NilObjectID, errors.New("user ID is not of expected type")
	}

	return primitive.ObjectIDFromHex(userIDHex)
}

// CreateTodo godoc
// @Summary      Create a new todo
// @Description  Adds a new todo item to the current user's list
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        todo body models.CreateTodoDTO true "Todo Create Object"
// @Success      201  {object}  models.Todo
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /todos [post]
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var dto models.CreateTodoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// now := primitive.NewDateTimeFromTime(time.Now())
	now := time.Now()
	newTodo := models.Todo{
		UserID:      userID,
		Title:       dto.Title,
		Description: dto.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := h.collection.InsertOne(context.Background(), newTodo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	newTodo.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, newTodo)
}

// GetAllTodos godoc
// @Summary      Get all todos for the current user
// @Description  Retrieves a list of all todo items belonging to the user
// @Tags         todos
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}  models.Todo
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /todos [get]
func (h *TodoHandler) GetAllTodos(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		println("ERROR: Failed to get user ID from context:", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	println("User ID:", userID.Hex())

	var todos []models.Todo
	filter := bson.M{"userId": userID}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := h.collection.Find(context.Background(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todos"})
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &todos); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode todos"})
		return
	}

	if todos == nil {
		todos = []models.Todo{}
	}

	c.JSON(http.StatusOK, todos)
}

// GetTodoByID godoc
// @Summary      Get a single todo by ID
// @Description  Retrieves a specific todo item by its ID, if it belongs to the user
// @Tags         todos
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path string true "Todo ID"
// @Success      200  {object}  models.Todo
// @Failure      400  {object}  map[string]string "Invalid ID format"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      404  {object}  map[string]string "Todo not found"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /todos/{id} [get]
func (h *TodoHandler) GetTodoByID(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var todo models.Todo
	filter := bson.M{"_id": id, "userId": userID}
	err = h.collection.FindOne(context.Background(), filter).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found or you don't have permission"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todo"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// UpdateTodo godoc
// @Summary      Update a todo
// @Description  Updates the details of a specific todo item
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path string true "Todo ID"
// @Param        todo body models.UpdateTodoDTO true "Todo Update Object"
// @Success      200  {object}  map[string]string "{'message': 'Todo updated successfully'}"
// @Failure      400  {object}  map[string]string "Invalid input or ID format"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      404  {object}  map[string]string "Todo not found"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /todos/{id} [put]
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var dto models.UpdateTodoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	update := bson.D{}
	if dto.Title != nil {
		update = append(update, bson.E{Key: "title", Value: *dto.Title})
	}
	if dto.Description != nil {
		update = append(update, bson.E{Key: "description", Value: *dto.Description})
	}
	if dto.Completed != nil {
		update = append(update, bson.E{Key: "completed", Value: *dto.Completed})
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No update fields provided"})
		return
	}
	update = append(update, bson.E{Key: "updatedAt", Value: primitive.NewDateTimeFromTime(time.Now())})

	filter := bson.M{"_id": id, "userId": userID}
	result, err := h.collection.UpdateOne(context.Background(), filter, bson.D{{Key: "$set", Value: update}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found or you don't have permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

// DeleteTodo godoc
// @Summary      Delete a todo
// @Description  Deletes a specific todo item by its ID
// @Tags         todos
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path string true "Todo ID"
// @Success      200  {object}  map[string]string "{'message': 'Todo deleted successfully'}"
// @Failure      400  {object}  map[string]string "Invalid ID format"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      404  {object}  map[string]string "Todo not found"
// @Failure      500  {object}  map[string]string "Server error"
// @Router       /todos/{id} [delete]
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	filter := bson.M{"_id": id, "userId": userID}
	result, err := h.collection.DeleteOne(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found or you don't have permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}
