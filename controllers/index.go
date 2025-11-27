package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Index godoc
//
//	@Summary		Welcome Message
//	@Description	Returns a welcome message for the API.
//	@Tags			General
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/ [get]
func Index(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Welcome to Vishwakarma Setu API",
	})
}

// HealthCheck godoc
//
//	@Summary		Health Check
//	@Description	Checks if the backend server is running.
//	@Tags			General
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "OK",
	})
}

// NotFound godoc
//
//	@Summary	Not Found Handler
//	@Tags		Errors
//	@Produce	json
//	@Success	404	{object}	map[string]string
//	@Router		/not-found [get]
func NotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, map[string]string{
		"error": "Resource not found",
	})
}

// InternalServerError godoc
//
//	@Summary	Internal Server Error Handler
//	@Tags		Errors
//	@Produce	json
//	@Success	500	{object}	map[string]string
//	@Router		/internal-server-error [get]
func InternalServerError(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Internal server error",
	})
}

// BadRequest godoc
//
//	@Summary	Bad Request Handler
//	@Tags		Errors
//	@Produce	json
//	@Success	400	{object}	map[string]string
//	@Router		/bad-request [get]
func BadRequest(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{
		"error": "Bad request",
	})
}

// Unauthorized godoc
//
//	@Summary	Unauthorized Handler
//	@Tags		Errors
//	@Produce	json
//	@Success	401	{object}	map[string]string
//	@Router		/unauthorized [get]
func Unauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{
		"error": "Unauthorized access",
	})
}

// Forbidden godoc
//
//	@Summary	Forbidden Handler
//	@Tags		Errors
//	@Produce	json
//	@Success	403	{object}	map[string]string
//	@Router		/forbidden [get]
func Forbidden(c echo.Context) error {
	return c.JSON(http.StatusForbidden, map[string]string{
		"error": "Forbidden access",
	})
}
