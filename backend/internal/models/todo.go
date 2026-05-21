package models

import (
	"time" // Import the standard time package
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Todo represents a single task in the ToDo list.
type Todo struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID      primitive.ObjectID `bson:"userId" json:"userId"` // Link to the User
	Title       string             `bson:"title" json:"title" binding:"required"`
	Description string             `bson:"description" json:"description"`
	Completed   bool               `bson:"completed" json:"completed"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// CreateTodoDTO is the Data Transfer Object for creating a new Todo.
type CreateTodoDTO struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// UpdateTodoDTO is the Data Transfer Object for updating an existing Todo.
type UpdateTodoDTO struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

