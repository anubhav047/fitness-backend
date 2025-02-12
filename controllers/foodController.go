package controllers

import (
	"context"
	"fitness-backend/models"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FoodController struct {
	db *mongo.Database
}

func NewFoodController(db *mongo.Database) *FoodController {
	return &FoodController{db: db}
}

func (fc *FoodController) AddFoodItem(c echo.Context) error {
	userId, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userObjID, _ := primitive.ObjectIDFromHex(userId)

	var foodItem models.FoodItem
	if err := c.Bind(&foodItem); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	foodItem.ID = primitive.NewObjectID()
	foodItem.ConsumedAt = time.Now()

	collection := fc.db.Collection("FoodConsumed")
	ctx := context.Background()

	filter := bson.M{"userId": userObjID}
	update := bson.M{
		"$push": bson.M{"foodItems": foodItem},
		"$set":  bson.M{"updatedAt": time.Now()},
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"userId":    userObjID,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add food item"})
	}

	return c.JSON(http.StatusCreated, foodItem)
}

func (fc *FoodController) GetUserFoodItems(c echo.Context) error {
	userId, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	userObjID, _ := primitive.ObjectIDFromHex(userId)

	collection := fc.db.Collection("foodConsumed")
	ctx := context.Background()

	var foodConsumed models.FoodConsumed
	err := collection.FindOne(ctx, bson.M{"userId": userObjID}).Decode(&foodConsumed)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusOK, models.FoodConsumed{FoodItems: []models.FoodItem{}})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch food items"})
	}

	return c.JSON(http.StatusOK, foodConsumed)
}

func (fc *FoodController) DeleteFoodItem(c echo.Context) error {
	userId, ok := c.Get("user_id").(string)
	if !ok {
		fmt.Println("user_id is null")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	foodItemId := c.Param("id")

	userObjID, _ := primitive.ObjectIDFromHex(userId)
	foodItemObjID, err := primitive.ObjectIDFromHex(foodItemId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid food item ID"})
	}

	collection := fc.db.Collection("foodConsumed")
	ctx := context.Background()

	update := bson.M{
		"$pull": bson.M{"foodItems": bson.M{"_id": foodItemObjID}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"userId": userObjID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete food item"})
	}

	if result.ModifiedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Food item not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Food item deleted successfully"})
}
