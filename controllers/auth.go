package controllers

import (
	"fitness-backend/models"
	"fitness-backend/utils"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	collection *mongo.Collection
}

func NewAuthController(db *mongo.Database) *AuthController {
	return &AuthController{
		collection: db.Collection("users"),
	}
}

func (ac *AuthController) SignUp(c echo.Context) error {
	// Check content type before binding
	contentType := c.Request().Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Content-Type must be application/json")
	}

	var user models.User
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var existingUser models.User
	err := ac.collection.FindOne(c.Request().Context(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "User with this email already exists"})
	} else if err != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error checking existing user"})
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	result, err := ac.collection.InsertOne(c.Request().Context(), user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error creating user"})
	}

	token, _ := utils.GenerateToken(result.InsertedID.(primitive.ObjectID).Hex())
	return c.JSON(http.StatusCreated, map[string]string{"token": token})
}

func (ac *AuthController) Login(c echo.Context) error {
	// Check content type before binding
	contentType := c.Request().Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Content-Type must be application/json")
	}

	var loginUser models.User
	if err := c.Bind(&loginUser); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var user models.User
	err := ac.collection.FindOne(c.Request().Context(), bson.M{"email": loginUser.Email}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	token, _ := utils.GenerateToken(user.ID.Hex())
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (ac *AuthController) GetUserDetails(c echo.Context) error {
	userIDString, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Unauthorized"))
	}
	objectID, err := primitive.ObjectIDFromHex(userIDString)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	var user models.User
	err = ac.collection.FindOne(c.Request().Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching user details")
	}

	// Remove sensitive information
	user.Password = ""

	return c.JSON(http.StatusOK, user)
}
