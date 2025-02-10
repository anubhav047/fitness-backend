package middleware

import (
	"fitness-backend/utils"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header required"})
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token format"})
		}

		token, err := utils.VerifyToken(bearerToken[1])
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])
		return next(c)
	}
}
