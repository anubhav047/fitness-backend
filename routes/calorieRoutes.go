package routes

import (
	"fitness-backend/middleware"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegisterRoutes sets up the application routes
func RegisterCalorieRoutes(e *echo.Echo, db *mongo.Database) {
	// Controllers

	// Protected API routes
	api := e.Group("/api")
	api.Use(middleware.AuthMiddleware)

}
