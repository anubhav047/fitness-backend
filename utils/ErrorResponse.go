package utils

import "github.com/labstack/echo/v4"

// ErrorResponse creates a JSON error response

func ErrorResponse(message string) echo.Map {

	return echo.Map{"error": message}

}

// SuccessResponse creates a JSON success response

func SuccessResponse(message string) echo.Map {

	return echo.Map{"message": message}

}
