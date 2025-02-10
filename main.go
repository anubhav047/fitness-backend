package main

import (
	"context"
	"fitness-backend/controllers"
	"fitness-backend/middleware"
	"fitness-backend/utils"
	"log"

	"github.com/labstack/echo/v4"
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

	// Auth routes
	auth := e.Group("/auth")
	auth.POST("/signup", authController.SignUp)
	auth.POST("/login", authController.Login)

	// Protected routes
	api := e.Group("/api")
	api.Use(middleware.AuthMiddleware)
	// Add protected routes here

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
