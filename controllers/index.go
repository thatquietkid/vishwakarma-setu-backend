package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Index(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Welcome to the Student Dashboard API",
	})
}
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "OK",
	})
}
func NotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, map[string]string{
		"error": "Resource not found",
	})
}
func InternalServerError(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Internal server error",
	})
}
func BadRequest(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{
		"error": "Bad request",
	})
}
func Unauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{
		"error": "Unauthorized access",
	})
}
func Forbidden(c echo.Context) error {
	return c.JSON(http.StatusForbidden, map[string]string{
		"error": "Forbidden access",
	})
}