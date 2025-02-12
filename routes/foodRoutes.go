package routes

import (
	"fitness-backend/controllers"
	"fitness-backend/middleware"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegisterRoutes sets up the application routes
func RegisterFoodRoutes(e *echo.Echo, db *mongo.Database) {
	// Controllers
	foodController := controllers.NewFoodController(db)

	// Protected API routes
	api := e.Group("/api")
	api.Use(middleware.AuthMiddleware)

	// Food routes
	api.POST("/food", foodController.AddFoodItem)
	api.GET("/food", foodController.GetUserFoodItems)
	api.DELETE("/food/:id", foodController.DeleteFoodItem)
}
