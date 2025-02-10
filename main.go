package main

import (
	"context"
	"fitness-backend/controllers"
	"fitness-backend/middleware"
	"fitness-backend/utils"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables
	utils.LoadEnv()

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(utils.GetEnvVariable("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("fitness")
	authController := controllers.NewAuthController(db)

	// Setup Gin router
	r := gin.Default()

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/signup", authController.SignUp)
		auth.POST("/login", authController.Login)
	}

	// Protected routes example
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// Add protected routes here
	}

	// Get port from environment variable or use default
	port := utils.GetEnvVariable("PORT")
	if port == "" {
		port = ":8080"
	} else {
		port = ":" + port
	}

	// Start server with error handling
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
