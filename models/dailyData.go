// models/dailyData.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DailyDataCollection represents the daily data for a user.
// The Goals field is dynamic and can hold different goal types (e.g., ExerciseGoal or FoodGoal).
type DailyDataCollection struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"` // Unique identifier
	UserID primitive.ObjectID `bson:"userId" json:"userId"`    // Reference to the user
	Date   time.Time          `bson:"date" json:"date"`        // The date of the record
	Type   string             `bson:"type" json:"type"`        // "exercise" or "calories"
	Goals  []interface{}      `bson:"goals" json:"goals"`      // Dynamic field for storing different goal types
}

// ExerciseGoal represents a goal for an exercise.
type ExerciseGoal struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`            // Unique identifier
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`               // Reference to the user
	ExerciseID    primitive.ObjectID `bson:"exerciseId" json:"exerciseId"`       // Reference to the exercise
	GoalName      string             `bson:"goalName" json:"goalName"`           // Name or description of the goal
	Type          string             `bson:"type" json:"type"`                   // "reps", "mins", or "kms"
	GoalValue     float64            `bson:"goalValue" json:"goalValue"`         // Target value for the goal
	ProgressValue float64            `bson:"progressValue" json:"progressValue"` // Current progress towards the goal
	Comments      string             `bson:"comments" json:"comments"`           // Additional comments
	IsActive      bool               `bson:"isActive" json:"isActive"`           // Indicates if the goal is active
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`         // Creation timestamp
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`         // Last update timestamp
}

// NutritionGoal represents a goal for nutrition intake.(WATER,CALORIES,CUSTOM GOALS)
// all goals whose goalName is not water or calories are custom goals
type NutritionGoal struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`            // Unique identifier
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`               // Reference to the user
	GoalName      string             `bson:"goalName" json:"goalName"`           // Name of the goal (calories/water)
	Type          string             `bson:"type" json:"type"`                   // "kcal" "L" or "g"
	GoalValue     float64            `bson:"goalValue" json:"goalValue"`         // Target value
	ProgressValue float64            `bson:"progressValue" json:"progressValue"` // Current intake value
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`         // Creation timestamp
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`         // Last update timestamp
}

// WeightEntry represents a single weight measurement
type WeightEntry struct {
	Value float64   `bson:"value" json:"value"` // Weight value
	Date  time.Time `bson:"date" json:"date"`   // Date of measurement
}

// WeightGoal represents a goal for weight management
type WeightGoal struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`          // Unique identifier
	UserID       primitive.ObjectID `bson:"userId" json:"userId"`             // Reference to the user
	GoalValue    float64            `bson:"goalValue" json:"goalValue"`       // Target weight
	CurrentValue float64            `bson:"currentValue" json:"currentValue"` // Current weight
	Unit         string             `bson:"unit" json:"unit"`                 // kg or lbs
	Entries      []WeightEntry      `bson:"entries" json:"entries"`           // History of weight entries
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`       // Creation timestamp
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`       // Last update timestamp
}
