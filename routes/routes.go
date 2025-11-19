package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/vishwakarma-setu-backend/controllers"
	"github.com/vishwakarma-setu-backend/middleware"
)

func RegisterRoutes(e *echo.Echo) {
	e.GET("/", controllers.Index)
	e.GET("/health", controllers.HealthCheck)

	// Error handling routes
	e.GET("/not-found", controllers.NotFound)
	e.GET("/internal-server-error", controllers.InternalServerError)
	e.GET("/bad-request", controllers.BadRequest)
	e.GET("/unauthorized", controllers.Unauthorized)
	e.GET("/forbidden", controllers.Forbidden)

	api := e.Group("/api")

	api.GET("/machines", controllers.GetAllListings)
	api.GET("/machines/:id", controllers.GetListingByID)

	// Protected Routes (Auth Required)
	// We group these routes and apply the JWTMiddleware
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware())

	protected.POST("/machines", controllers.CreateListing)
	protected.PUT("/machines/:id", controllers.UpdateListing)
	protected.DELETE("/machines/:id", controllers.DeleteListing)
}