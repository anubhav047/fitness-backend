package routes

import (
	"fitness-backend/controllers"
	"fitness-backend/middleware"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegisterRoutes sets up the application routes
func RegisterRoutes(e *echo.Echo, db *mongo.Database) {
	// Controllers
	exerciseGuideController := controllers.NewExerciseGuideController(db)

	// Protected API routes
	api := e.Group("/api")
	api.Use(middleware.AuthMiddleware)

	// Exercise Guide Routes
	exercise := api.Group("/exercises")
	exercise.GET("", exerciseGuideController.GetExercises)

	// Category-based routes
	categoryGroup := exercise.Group("/category")
	categoryGroup.GET("/:category", exerciseGuideController.GetExercisesByCategory)

	// Exercise CRUD routes
	exercise.GET("/:id", exerciseGuideController.GetExerciseByID)
	exercise.POST("", exerciseGuideController.CreateExercise)
	exercise.DELETE("/:id", exerciseGuideController.DeleteExerciseByID)
}
