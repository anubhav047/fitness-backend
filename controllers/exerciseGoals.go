package controllers

import (
	"context"
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
	Collection         *mongo.Collection
	ExerciseCollection *mongo.Collection // Add this line
}

// Modify NewGoalController
func NewGoalController(db *mongo.Database) *GoalController {
	return &GoalController{
		Collection:         db.Collection("daily_data"),
		ExerciseCollection: db.Collection("exercise_guides"), // Add this line
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
		var goalMap map[string]interface{}
		if err := mapstructure.Decode(request.Goal, &goalMap); err != nil {
			fmt.Println("Decode error:", err) // Debugging line
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid exercise goal data"))
		}

		// Convert ExerciseID from string to primitive.ObjectID
		if exerciseIDStr, ok := goalMap["exerciseId"].(string); ok {
			exerciseID, err := primitive.ObjectIDFromHex(exerciseIDStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid exercise ID format"))
			}
			goalMap["exerciseId"] = exerciseID
		}

		var goal models.ExerciseGoal
		if err := mapstructure.Decode(goalMap, &goal); err != nil {
			fmt.Println("Decode error:", err) // Debugging line
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
	// Retrieve the updated goal
	var dailyData models.DailyDataCollection
	err = gc.Collection.FindOne(c.Request().Context(), filter).Decode(&dailyData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to retrieve updated goal"))
	}

	var updatedGoal interface{}
	for _, goal := range dailyData.Goals {
		switch g := goal.(type) {
		case primitive.D:
			goalMap := g.Map()
			if goalMap["_id"] == goalID {
				updatedGoal = goalMap
				break
			}
		default:
			fmt.Printf("Unknown goal type: %T\n", goal)
		}
	}

	if updatedGoal == nil {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse("Updated goal not found"))
	}

	return c.JSON(http.StatusOK, updatedGoal)

	// return c.JSON(http.StatusOK, utils.SuccessResponse("Goal updated successfully"))
}

// DeleteGoal removes a goal by ID, sets goalValue to 0, and deletes the document if both goalValue and progressValue are 0.
func (gc *GoalController) DeleteGoal(c echo.Context) error {
	// Get user ID
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Unauthorized"))
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid user ID format"))
	}

	// Get goal ID and date
	goalID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal ID format"))
	}

	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	// Debug log
	fmt.Printf("Searching for goal: UserID=%s, Date=%s, GoalID=%s\n",
		userID.Hex(), date.Format("2006-01-02"), goalID.Hex())

	// Create filter with both string and ObjectID possibilities
	filter := bson.M{
		"userId": userID,
		"date":   date,
		"goals._id": bson.M{
			"$in": []interface{}{goalID, goalID.Hex()},
		},
	}

	// First, verify goal exists
	var dailyData models.DailyDataCollection
	err = gc.Collection.FindOne(c.Request().Context(), filter).Decode(&dailyData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Database error"))
	}

	// Find the specific goal
	var targetGoal interface{}
	for _, goal := range dailyData.Goals {
		var goalMap bson.M
		switch g := goal.(type) {
		case primitive.D:
			goalMap = g.Map()
		case bson.M:
			goalMap = g
		default:
			continue
		}

		// Check both ObjectID and string formats
		if gID, ok := goalMap["_id"]; ok {
			switch id := gID.(type) {
			case primitive.ObjectID:
				if id == goalID {
					targetGoal = goalMap
					break
				}
			case string:
				if id == goalID.Hex() {
					targetGoal = goalMap
					break
				}
			}
		}
	}

	if targetGoal == nil {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found in document"))
	}

	// Set goalValue to 0 in the database
	update := bson.M{
		"$set": bson.M{"goals.$.goalValue": 0},
	}

	updateResult, err := gc.Collection.UpdateOne(c.Request().Context(), filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update goal"))
	}

	if updateResult.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found for update"))
	}

	// Update our in-memory targetGoal to reflect the change
	goalMap := targetGoal.(bson.M)
	goalMap["goalValue"] = 0

	// Retrieve values from the updated targetGoal
	goalValue, _ := utils.ConvertToFloat(goalMap["goalValue"])
	progressValue, _ := utils.ConvertToFloat(goalMap["progressValue"])

	// Check if both values are 0; if so, delete the goal
	if goalValue == 0 && progressValue == 0 {
		deleteFilter := bson.M{
			"userId": userID,
			"date":   date,
		}
		deleteUpdate := bson.M{
			"$pull": bson.M{
				"goals": bson.M{
					"_id": bson.M{
						"$in": []interface{}{goalID, goalID.Hex()},
					},
				},
			},
		}

		deleteResult, err := gc.Collection.UpdateOne(
			c.Request().Context(),
			deleteFilter,
			deleteUpdate,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to delete goal"))
		}

		if deleteResult.ModifiedCount > 0 {
			return c.JSON(http.StatusOK, utils.SuccessResponse("Goal deleted successfully"))
		}
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse("Goal value set to 0"))
}

func (gc *GoalController) UpsertGoalByName(c echo.Context) error {
	// Step 1: Validate request context
	ctx := c.Request().Context()
	if ctx.Err() != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Request context canceled"))
	}

	// Step 2: Authorization check
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Unauthorized"))
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid user ID"))
	}

	// Step 3: Parse parameters
	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	goalName := c.Param("goalName")
	if goalName == "" {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Goal name is required"))
	}

	// Step 4: Parse and validate request body
	var request struct {
		Type string      `json:"type"`
		Goal interface{} `json:"goal"`
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid request format"))
	}

	if request.Type == "" || request.Goal == nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Type and goal are required"))
	}

	// Step 5: Check if goal exists
	filter := bson.M{
		"userId":         userID,
		"date":           date,
		"goals.goalName": goalName,
	}

	var dailyData models.DailyDataCollection
	err = gc.Collection.FindOne(ctx, filter).Decode(&dailyData)

	// Step 6: Handle create/update logic
	if err == mongo.ErrNoDocuments {
		// Create new goal
		goalID := primitive.NewObjectID()
		var goal interface{}

		switch request.Type {
		case "exercise":
			goal, err = gc.createExerciseGoal(request.Goal, goalID, userID)
			if err != nil {
				return c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
			}
		default:
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Unsupported goal type"))
		}

		update := bson.M{"$push": bson.M{"goals": goal}}
		opts := options.Update().SetUpsert(true)

		_, err = gc.Collection.UpdateOne(ctx,
			bson.M{"userId": userID, "date": date},
			update,
			opts)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to create goal"))
		}

		return c.JSON(http.StatusCreated, goal)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Database error"))
	}

	// Step 7: Update existing goal
	goalMap, err := convertGoalToMap(request.Goal)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal data"))
	}

	goalMap["updatedAt"] = time.Now()

	update := bson.M{}
	for key, value := range goalMap {
		update["goals.$."+key] = value
	}

	updateResult, err := gc.Collection.UpdateOne(
		ctx,
		filter,
		bson.M{"$set": update},
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update goal"))
	}

	if updateResult.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
	}

	return c.JSON(http.StatusOK, goalMap)
}

// Helper functions
func (gc *GoalController) createExerciseGoal(goalData interface{}, goalID, userID primitive.ObjectID) (models.ExerciseGoal, error) {
	var goalMap map[string]interface{}
	if err := mapstructure.Decode(goalData, &goalMap); err != nil {
		return models.ExerciseGoal{}, fmt.Errorf("invalid goal data format")
	}

	// Debug print
	fmt.Printf("Goal Map: %+v\n", goalMap)

	// Get goal name
	goalName, ok := goalMap["goalName"].(string)
	if !ok {
		return models.ExerciseGoal{}, fmt.Errorf("invalid goalName format")
	}

	// Get exercise ID from exercise_guides collection based on goal name
	exerciseID := gc.getExerciseIDByName(context.Background(), goalName)

	// Rest of the validation
	goalType, ok := goalMap["type"].(string)
	if !ok {
		return models.ExerciseGoal{}, fmt.Errorf("invalid type format")
	}

	goalValue, ok := goalMap["goalValue"].(float64)
	if !ok {
		if gv, ok := goalMap["goalValue"].(int); ok {
			goalValue = float64(gv)
		} else {
			return models.ExerciseGoal{}, fmt.Errorf("invalid goalValue format")
		}
	}

	progressValue, ok := goalMap["progressValue"].(float64)
	if !ok {
		if pv, ok := goalMap["progressValue"].(int); ok {
			progressValue = float64(pv)
		} else {
			return models.ExerciseGoal{}, fmt.Errorf("invalid progressValue format")
		}
	}

	comments, ok := goalMap["comments"].(string)
	if !ok {
		comments = "" // Default to empty string if not provided
	}

	isActive, ok := goalMap["isActive"].(bool)
	if !ok {
		isActive = true // Default to true if not provided
	}

	return models.ExerciseGoal{
		ID:            goalID,
		UserID:        userID,
		ExerciseID:    exerciseID, // Use the found or empty exercise ID
		GoalName:      goalName,
		Type:          goalType,
		GoalValue:     goalValue,
		ProgressValue: progressValue,
		Comments:      comments,
		IsActive:      isActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

func convertGoalToMap(goalData interface{}) (map[string]interface{}, error) {
	var goalMap map[string]interface{}
	if err := mapstructure.Decode(goalData, &goalMap); err != nil {
		return nil, fmt.Errorf("invalid goal data format")
	}
	return goalMap, nil
}

func (gc *GoalController) getExerciseIDByName(ctx context.Context, name string) primitive.ObjectID {
	var exercise models.Exercise
	err := gc.ExerciseCollection.FindOne(ctx, bson.M{"name": name}).Decode(&exercise)
	if err != nil {
		return primitive.NilObjectID
	}
	return exercise.ID
}
