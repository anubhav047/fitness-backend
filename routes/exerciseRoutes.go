package routes

import (
	"fitness-backend/controllers"
	"fitness-backend/middleware"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegisterRoutes sets up the application routes
func RegisterExerciseRoutes(e *echo.Echo, db *mongo.Database) {
	// Controllers
	exerciseGuideController := controllers.NewExerciseGuideController(db)
	goalController := controllers.NewGoalController(db)
	progressController := controllers.NewProgressController(db)

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

	// Goal Management Routes
	goals := api.Group("/goals")
	goals.POST("/:date", goalController.CreateGoal)           // Create
	goals.GET("/:date/:id", goalController.GetGoal)           // Specific GET
	goals.PATCH("/:date/:id", goalController.UpdateGoal)      // Specific PATCH
	goals.GET("/:date/active", goalController.GetActiveGoals) // Get all active
	goals.GET("/:date", goalController.GetAllGoals)           // Get all
	goals.DELETE("/:date/:id", goalController.DeleteGoal)     // Delete

	// Progress Management Routes
	progress := api.Group("/progress")
	// progress.POST("/:date", goalController.CreateGoal)         // Create goal with goalvalue 0
	progress.GET("/:date", progressController.GetProgress)           // Get all progress for a user for a date
	progress.PATCH("/:id", progressController.UpdateProgress)        // Update progress entry
	progress.DELETE("/:date/:id", progressController.DeleteProgress) // Delete progress entry
}
