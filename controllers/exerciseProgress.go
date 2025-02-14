package controllers

import (
	"fitness-backend/models"
	"fitness-backend/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProgressController struct {
	Collection *mongo.Collection
}

func NewProgressController(db *mongo.Database) *ProgressController {
	return &ProgressController{
		Collection: db.Collection("daily_data"),
	}
}

// GetProgress retrieves the sum of progressValues and goalValues for a specific date
func (pc *ProgressController) GetProgress(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	fmt.Println(userIDString)
	if !ok {
		fmt.Println("user_id not found")
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Unauthorized"))
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Invalid user ID format"))
	}

	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	filter := bson.M{"userId": userID, "date": date}
	var dailyData models.DailyDataCollection
	err = pc.Collection.FindOne(c.Request().Context(), filter).Decode(&dailyData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse("No progress found for this date"))
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error retrieving progress"))
	}

	var totalProgress, totalGoal float64
	for _, goal := range dailyData.Goals {
		var goalMap bson.M
		switch g := goal.(type) {
		case primitive.D:
			bsonBytes, err := bson.Marshal(g)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error marshaling goal"))
			}
			err = bson.Unmarshal(bsonBytes, &goalMap)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error unmarshaling goal"))
			}
		case bson.M:
			goalMap = g
		default:
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Unknown goal type"))
		}

		progressValue, _ := goalMap["progressValue"].(float64)
		goalValue, _ := goalMap["goalValue"].(float64)

		// Only include progress (and goal) if goalValue is greater than 0
		if goalValue > 0 {
			totalProgress += progressValue
			totalGoal += goalValue
		}
	}

	return c.JSON(http.StatusOK, bson.M{"totalProgress": totalProgress, "totalGoal": totalGoal})
}

// DeleteProgress sets progressValue to 0 and deletes goal if both values are 0
func (pc *ProgressController) DeleteProgress(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Unauthorized"))
	}

	userID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Invalid user ID format"))
	}

	goalID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid goal ID format"))
	}

	date, err := time.Parse("2006-01-02", c.Param("date"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid date format"))
	}

	filter := bson.M{"userId": userID, "date": date, "goals._id": bson.M{
		"$in": []interface{}{goalID, goalID.Hex()},
	}}
	update := bson.M{"$set": bson.M{"goals.$.progressValue": 0}}

	result, err := pc.Collection.UpdateOne(c.Request().Context(), filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update progress"))
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse("Goal not found"))
	}

	// Retrieve updated document
	var dailyData models.DailyDataCollection
	err = pc.Collection.FindOne(c.Request().Context(), filter).Decode(&dailyData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to retrieve goal"))
	}

	// Check if goal should be deleted
	for _, goal := range dailyData.Goals {
		var goalMap bson.M

		switch g := goal.(type) {
		case primitive.D:
			bsonBytes, err := bson.Marshal(g)
			if err != nil {
				continue
			}
			if err := bson.Unmarshal(bsonBytes, &goalMap); err != nil {
				continue
			}
		case bson.M:
			goalMap = g
		default:
			continue
		}

		// Handle both string and ObjectID types for _id
		var matchFound bool
		switch gID := goalMap["_id"].(type) {
		case primitive.ObjectID:
			matchFound = gID == goalID
		case string:
			matchFound = gID == goalID.Hex()
		}

		if matchFound {
			goalValue, _ := utils.ConvertToFloat(goalMap["goalValue"])
			progressValue, _ := utils.ConvertToFloat(goalMap["progressValue"])

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

				_, err := pc.Collection.UpdateOne(
					c.Request().Context(),
					deleteFilter,
					deleteUpdate,
				)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to delete goal"))
				}
				return c.JSON(http.StatusOK, utils.SuccessResponse("Goal deleted successfully"))
			}
			break
		}
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse("Progress set to 0"))
}
