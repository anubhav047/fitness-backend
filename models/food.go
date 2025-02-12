package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FoodItem represents a single food item with its nutritional information
type FoodItem struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                string             `bson:"name" json:"name"`
	ConsumedAt          time.Time          `bson:"consumedAt" json:"consumedAt"`
	ServingSizeG        float64            `bson:"serving_size_g" json:"serving_size_g"`
	Calories            float64            `bson:"calories" json:"calories"`
	FatTotalG           float64            `bson:"fat_total_g" json:"fat_total_g"`
	FatSaturatedG       float64            `bson:"fat_saturated_g" json:"fat_saturated_g"`
	ProteinG            float64            `bson:"protein_g" json:"protein_g"`
	SodiumMg            float64            `bson:"sodium_mg" json:"sodium_mg"`
	PotassiumMg         float64            `bson:"potassium_mg" json:"potassium_mg"`
	CholesterolMg       float64            `bson:"cholesterol_mg" json:"cholesterol_mg"`
	CarbohydratesTotalG float64            `bson:"carbohydrates_total_g" json:"carbohydrates_total_g"`
	FiberG              float64            `bson:"fiber_g" json:"fiber_g"`
	SugarG              float64            `bson:"sugar_g" json:"sugar_g"`
}

// FoodConsumed represents the collection of food items consumed by a user
type FoodConsumed struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	FoodItems []FoodItem         `bson:"foodItems" json:"foodItems"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
