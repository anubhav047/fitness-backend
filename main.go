package main

import (
	"context"
	"fitness-backend/controllers"
	"fitness-backend/routes"
	"fitness-backend/utils"
	"log"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	utils.LoadEnv()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(utils.GetEnvVariable("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("fitness")
	authController := controllers.NewAuthController(db)

	// Setup Echo
	e := echo.New()

	// Get frontend origins from environment variable
	frontendOrigins := os.Getenv("FRONTEND_ORIGINS")
	if frontendOrigins == "" {
		log.Fatal("FRONTEND_ORIGINS environment variable is not set")
	}
	origins := strings.Split(frontendOrigins, ",")

	// Apply CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins, // Use the frontend origins from the environment variable
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
	}))

	// Auth routes
	auth := e.Group("/auth")
	auth.POST("/signup", authController.SignUp)
	auth.POST("/login", authController.Login)

	//Routes for exercises
	routes.RegisterExerciseRoutes(e, db)
	//Routes for calories
	routes.RegisterFoodRoutes(e, db)

	port := utils.GetEnvVariable("PORT")
	if port == "" {
		port = ":8080"
	} else {
		port = ":" + port
	}

	if err := e.Start(port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
