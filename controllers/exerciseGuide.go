package controllers

import (
	"context"
	"fitness-backend/models"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ExerciseGuideController struct {
	collection *mongo.Collection
}

// NewExerciseGuideController initializes a new instance of ExerciseGuideController
func NewExerciseGuideController(db *mongo.Database) *ExerciseGuideController {
	return &ExerciseGuideController{
		collection: db.Collection("exercise_guides"),
	}
}

// GetExercises retrieves all exercise guides from the database
func (ec *ExerciseGuideController) GetExercises(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exercises []models.Exercise
	cursor, err := ec.collection.Find(ctx, bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch exercises"})
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var exercise models.Exercise
		if err := cursor.Decode(&exercise); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error decoding exercise"})
		}
		exercises = append(exercises, exercise)
	}

	return c.JSON(http.StatusOK, exercises)
}

// GetExercisesByCategory retrieves all exercise guides for a specific category
func (ec *ExerciseGuideController) GetExercisesByCategory(c echo.Context) error {
	category := c.Param("category")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exercises []models.Exercise
	cursor, err := ec.collection.Find(ctx, bson.M{"category": category})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch exercises"})
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var exercise models.Exercise
		if err := cursor.Decode(&exercise); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error decoding exercise"})
		}
		exercises = append(exercises, exercise)
	}

	if err := cursor.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Cursor error"})
	}

	if len(exercises) == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "No exercises found for this category"})
	}

	return c.JSON(http.StatusOK, exercises)
}

// GetExerciseByID retrieves a single exercise guide by its ID
func (ec *ExerciseGuideController) GetExerciseByID(c echo.Context) error {
	exerciseID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(exerciseID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid exercise ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exercise models.Exercise
	err = ec.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&exercise)
	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Exercise not found"})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch exercise"})
	}

	return c.JSON(http.StatusOK, exercise)
}

// CreateExercise adds a new exercise guide to the database
func (ec *ExerciseGuideController) CreateExercise(c echo.Context) error {
	var newExercise models.Exercise

	// Bind and validate input
	if err := c.Bind(&newExercise); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input", "details": err.Error()})
	}

	// Validate required fields
	if newExercise.Name == "" || newExercise.Category == "" || len(newExercise.MainMuscles) == 0 ||
		newExercise.Difficulty == "" || len(newExercise.Benefits) == 0 || len(newExercise.Steps) == 0 ||
		len(newExercise.Tips) == 0 || newExercise.EstimatedTime == "" || newExercise.VideoURL == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "All fields are required"})
	}

	// Assign new ObjectID
	newExercise.ID = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert into MongoDB
	result, err := ec.collection.InsertOne(ctx, newExercise)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create exercise", "details": err.Error()})
	}

	// Return the created exercise
	return c.JSON(http.StatusCreated, echo.Map{
		"message":  "Exercise created successfully",
		"exercise": newExercise,
		"id":       result.InsertedID,
	})
}

// DeleteExerciseByID removes an exercise guide by its ID
func (ec *ExerciseGuideController) DeleteExerciseByID(c echo.Context) error {
	exerciseID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(exerciseID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid exercise ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := ec.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete exercise"})
	}

	if result.DeletedCount == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Exercise not found"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Exercise deleted successfully"})
}
