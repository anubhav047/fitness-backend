package utils

import (
	"time"

	"github.com/golang-jwt/jwt"
)

func GenerateToken(userId string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userId

	// Set expiration claim to 24 hours from now
	// "exp" is a standard JWT claim for expiration time
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()

	jwtSecret := []byte(GetEnvVariable("JWT_SECRET"))
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	jwtSecret := []byte(GetEnvVariable("JWT_SECRET"))
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}
