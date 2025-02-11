package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Exercise struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name" validate:"required"`
	Category      string             `bson:"category" json:"category" validate:"required"`
	MainMuscles   []string           `bson:"mainMuscles" json:"mainMuscles" validate:"required"`
	Difficulty    string             `bson:"difficulty" json:"difficulty" validate:"required"`
	Benefits      []string           `bson:"benefits" json:"benefits" validate:"required"`
	Steps         []string           `bson:"steps" json:"steps" validate:"required"`
	Tips          []string           `bson:"tips" json:"tips" validate:"required"`
	EstimatedTime string             `bson:"estimatedTime" json:"estimatedTime" validate:"required"`
	VideoURL      string             `bson:"videoUrl" json:"videoUrl" validate:"required,url"`
}
