package controllers

import (
	"fitness-backend/models"
	"fitness-backend/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GoalController struct {
	Collection *mongo.Collection
}

func NewGoalController(db *mongo.Database) *GoalController {
	return &GoalController{
		Collection: db.Collection("daily_data"),
	}
}

// GetAllGoals retrieves all goals for a specific date
func (gc *GoalController) GetAllGoals(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		fmt.Println("Invalid user ID format")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	filter := bson.M{"userId": userID, "date": date}
	var dailyData models.DailyDataCollection
	err = gc.Collection.FindOne(c.Request().Context(), filter).Decode(&dailyData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse("No goals found for this date"))
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error retrieving goals"))
	}

	return c.JSON(http.StatusOK, dailyData.Goals)
}

// GetActiveGoals retrieves active goals for a specific date
func (gc *GoalController) GetActiveGoals(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		fmt.Println("Invalid user ID format")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	filter := bson.M{"userId": userID, "date": date, "goals.isActive": true}
	var dailyData models.DailyDataCollection
	err = gc.Collection.FindOne(c.Request().Context(), filter).Decode(&dailyData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse("No active goals found for this date"))
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error retrieving active goals"))
	}

	return c.JSON(http.StatusOK, dailyData.Goals)
}

// GetGoal retrieves a specific goal by ID for a specific date
func (gc *GoalController) GetGoal(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		fmt.Println("Invalid user ID format")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	goalID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal ID format"))
	}
	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	filter := bson.M{"userId": userID, "date": date}
	var dailyData models.DailyDataCollection
	err = gc.Collection.FindOne(c.Request().Context(), filter).Decode(&dailyData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error retrieving goal"))
	}

	// Iterate through the goals to find the one with the matching ID
	for _, goal := range dailyData.Goals {
		switch g := goal.(type) {
		case primitive.D:
			goalMap := g.Map()
			goalIDBson, ok := goalMap["_id"].(primitive.ObjectID)
			if ok && goalIDBson == goalID {
				// Try to decode into ExerciseGoal
				var exerciseGoal models.ExerciseGoal
				bsonBytes, err := bson.Marshal(goalMap)
				if err == nil {
					err = bson.Unmarshal(bsonBytes, &exerciseGoal)
					if err == nil {
						return c.JSON(http.StatusOK, exerciseGoal)
					}
				}

				// Try to decode into FoodGoal
				var foodGoal models.FoodGoal
				bsonBytes, err = bson.Marshal(goalMap)
				if err == nil {
					err = bson.Unmarshal(bsonBytes, &foodGoal)
					if err == nil {
						return c.JSON(http.StatusOK, foodGoal)
					}
				}
			}
		default:
			fmt.Printf("Unknown goal type: %T\n", goal)
		}
	}

	return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
}

// CreateGoal adds a new goal for a specific date
func (gc *GoalController) CreateGoal(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		fmt.Println("Invalid user ID format")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	var request struct {
		Type       string             `json:"type"`
		ExerciseID primitive.ObjectID `json:"exerciseId"`
		Goal       interface{}        `json:"goal"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid request body"))
	}

	goalID := primitive.NewObjectID()
	switch request.Type {
	case "exercise":
		var goal models.ExerciseGoal
		if err := mapstructure.Decode(request.Goal, &goal); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal data"))
		}
		goal.ID = goalID
		goal.UserID = userID
		goal.ExerciseID = request.ExerciseID
		goal.CreatedAt = time.Now()
		goal.UpdatedAt = time.Now()
		request.Goal = goal
	case "calories":
		var goal models.ExerciseGoal
		if err := mapstructure.Decode(request.Goal, &goal); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal data"))
		}
		goal.ID = goalID
		goal.UserID = userID
		goal.CreatedAt = time.Now()
		goal.UpdatedAt = time.Now()
		request.Goal = goal
	default:
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal type"))
	}

	update := bson.M{"$push": bson.M{"goals": request.Goal}}
	_, err = gc.Collection.UpdateOne(c.Request().Context(), bson.M{"userId": userID, "date": date}, update, options.Update().SetUpsert(true))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to create goal"))
	}

	return c.JSON(http.StatusCreated, request.Goal)
}

// UpdateGoal updates a specific goal by ID for a specific date
func (gc *GoalController) UpdateGoal(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		fmt.Println("Invalid user ID format")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	goalID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal ID format"))
	}
	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	var updateData map[string]interface{}
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid request body"))
	}

	// Construct the update document
	update := bson.M{}
	for key, value := range updateData {
		update["goals.$."+key] = value
	}

	// Perform the update
	filter := bson.M{"userId": userID, "date": date, "goals._id": goalID}
	updateResult, err := gc.Collection.UpdateOne(c.Request().Context(), filter, bson.M{"$set": update})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update goal"))
	}

	if updateResult.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse("Goal updated successfully"))
}

// DeleteGoal removes a goal by ID, sets goalValue to 0, and deletes the document if both goalValue and progressValue are 0.
func (gc *GoalController) DeleteGoal(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		fmt.Println("Invalid user ID format")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	goalID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal ID format"))
	}
	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	// Find the goal to check its current values
	filter := bson.M{"userId": userID, "date": date}
	var dailyData models.DailyDataCollection
	err = gc.Collection.FindOne(c.Request().Context(), filter).Decode(&dailyData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error retrieving goal"))
	}

	// Locate the specific goal within the dailyData.Goals slice
	var goalToDelete interface{}
	for _, g := range dailyData.Goals {
		switch goal := g.(type) { // Type assertion for each goal
		case primitive.D:
			goalMap := goal.Map()
			goalIDBson, ok := goalMap["_id"].(primitive.ObjectID)
			if ok && goalIDBson == goalID {
				goalToDelete = goalMap
				break
			}
		default:
			continue // Skip unknown goal types
		}
	}

	if goalToDelete == nil {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
	}

	// Update the document by removing the goal
	update := bson.M{
		"$pull": bson.M{"goals": bson.M{"_id": goalID}},
	}
	_, err = gc.Collection.UpdateOne(c.Request().Context(), filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to delete goal"))
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse("Goal deleted successfully"))
}
