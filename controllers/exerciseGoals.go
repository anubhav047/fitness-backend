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
func convertToFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
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

				// Try to decode into NutritionGoal
				var nutritionGoal models.NutritionGoal
				bsonBytes, err = bson.Marshal(goalMap)
				if err == nil {
					err = bson.Unmarshal(bsonBytes, &nutritionGoal)
					if err == nil {
						return c.JSON(http.StatusOK, nutritionGoal)
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
		Type string      `json:"type"`
		Goal interface{} `json:"goal"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid request body"))
	}

	goalID := primitive.NewObjectID()
	switch request.Type {
	case "exercise":
		var goal models.ExerciseGoal
		if err := mapstructure.Decode(request.Goal, &goal); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid exercise goal data"))
		}
		goal.ID = goalID
		goal.UserID = userID
		goal.CreatedAt = time.Now()
		goal.UpdatedAt = time.Now()
		request.Goal = goal

	case "water", "calorie", "customgoal":
		var goal models.NutritionGoal
		if err := mapstructure.Decode(request.Goal, &goal); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid nutrition goal data"))
		}
		goal.ID = goalID
		goal.UserID = userID
		goal.CreatedAt = time.Now()
		goal.UpdatedAt = time.Now()
		request.Goal = goal

	case "weight":
		var goal models.WeightGoal
		if err := mapstructure.Decode(request.Goal, &goal); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid weight goal data"))
		}
		goal.ID = goalID
		goal.UserID = userID
		goal.CreatedAt = time.Now()
		goal.UpdatedAt = time.Now()
		// Initialize empty entries array if not provided
		if goal.Entries == nil {
			goal.Entries = []models.WeightEntry{}
		}
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
	// Step 1: Extract user ID from context
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Unauthorized"))
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		fmt.Println("Invalid user ID format")
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Internal Server Error"))
	}

	// Step 2: Extract Goal ID and Date
	goalID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal ID format"))
	}

	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	// Step 3: Update goalValue to 0
	filter := bson.M{"userId": userID, "date": date, "goals._id": goalID}
	update := bson.M{"$set": bson.M{"goals.$.goalValue": 0}}

	updateResult, err := gc.Collection.UpdateOne(c.Request().Context(), filter, update)
	if err != nil {
		fmt.Println("Failed to update goal value")
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update goal value"))
	}

	if updateResult.MatchedCount == 0 {
		fmt.Println("Goal not found")
		return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
	}

	// Step 4: Retrieve the updated document
	var dailyData models.DailyDataCollection
	err = gc.Collection.FindOne(c.Request().Context(), bson.M{"userId": userID, "date": date}).Decode(&dailyData)
	if err != nil {
		fmt.Println("Failed to retrieve goal")
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to retrieve goal"))
	}

	// Step 5: Locate the specific goal and check values
	for _, goal := range dailyData.Goals {
		var goalMap bson.M

		// Convert goal to bson.M properly
		switch g := goal.(type) {
		case primitive.D:
			bsonBytes, err := bson.Marshal(g)
			if err != nil {
				fmt.Println("Failed to marshal goal:", err)
				continue
			}
			err = bson.Unmarshal(bsonBytes, &goalMap)
			if err != nil {
				fmt.Println("Failed to unmarshal goal:", err)
				continue
			}
		case bson.M:
			goalMap = g
		default:
			fmt.Println("Skipping unknown goal type:", g)
			continue
		}

		// Debug output to verify goal extraction
		fmt.Printf("Converted Goal: %+v\n", goalMap)

		goalIDBson, ok := goalMap["_id"].(primitive.ObjectID)
		if ok && goalIDBson == goalID {
			goalValue, goalValueOk := convertToFloat(goalMap["goalValue"])
			progressValue, progressValueOk := convertToFloat(goalMap["progressValue"])

			fmt.Printf("Checking Goal Values -> goalValue: %v, progressValue: %v\n", goalValue, progressValue)

			// Step 6: Delete goal if both values are 0
			if goalValueOk && progressValueOk && goalValue == 0 && progressValue == 0 {
				deleteUpdate := bson.M{
					"$pull": bson.M{"goals": bson.M{"_id": goalIDBson}},
				}

				fmt.Printf("🛠️ Running Delete Query: %+v\n", deleteUpdate)

				deleteResult, err := gc.Collection.UpdateOne(c.Request().Context(), filter, deleteUpdate)
				if err != nil {
					fmt.Println("Failed to delete goal")
					return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to delete goal"))
				}

				// Log MongoDB deletion result
				fmt.Printf(" Delete Result: %+v\n", deleteResult)

				if deleteResult.ModifiedCount > 0 {
					fmt.Println(" Goal deleted successfully!")
					return c.JSON(http.StatusOK, utils.SuccessResponse("Goal deleted successfully"))
				} else {
					fmt.Println(" Goal deletion failed")
					return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Goal deletion failed"))
				}
			} else {
				fmt.Println("Goal value set to 0, but not deleted (progressValue not 0)")
			}
		}
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse("Goal value set to 0"))
}
