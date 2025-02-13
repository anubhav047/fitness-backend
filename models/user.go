package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string             `bson:"email" json:"email" validate:"required,email"`
	Name     string             `bson:"name" json:"name" validate:"required"`
	Password string             `bson:"password" json:"password" validate:"required"`
	Weight   float64            `bson:"weight" json:"weight"`
	Height   float64            `bson:"height" json:"height"`
}
