package routes

import (
	"github.com/labstack/echo/v4"
	swagger "github.com/swaggo/echo-swagger"
	"github.com/vishwakarma-setu-backend/controllers"
	"github.com/vishwakarma-setu-backend/middleware"
)

func RegisterRoutes(e *echo.Echo) {
	// Swagger Documentation Route
	e.GET("/swagger/*", swagger.WrapHandler)

	// API Group
	api := e.Group("/api")
	
	// Root & Utility Routes
	api.GET("/", controllers.Index)
	api.GET("/health", controllers.HealthCheck)

	// Error handling routes (Useful for testing standardized errors)
	api.GET("/not-found", controllers.NotFound)
	api.GET("/internal-server-error", controllers.InternalServerError)
	api.GET("/bad-request", controllers.BadRequest)
	api.GET("/unauthorized", controllers.Unauthorized)
	api.GET("/forbidden", controllers.Forbidden)

	// Public Machine Routes
	api.GET("/machines", controllers.GetAllListings)
	api.GET("/machines/:id", controllers.GetListingByID)

	// Protected Routes (Auth Required)
	// We group these routes and apply the JWTMiddleware
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware())

	// Machine Management
	protected.POST("/machines", controllers.CreateListing)
	protected.PUT("/machines/:id", controllers.UpdateListing)
	protected.DELETE("/machines/:id", controllers.DeleteListing)

	// Rental Management (Restored)
	protected.POST("/rentals", controllers.CreateRentalRequest)        // Renter books
	protected.GET("/rentals/my", controllers.GetMyRentals)             // Renter history
	protected.GET("/rentals/manage", controllers.GetOwnerRentals)      // Owner manages requests
	protected.PUT("/rentals/:id/status", controllers.UpdateRentalStatus) // Owner approves/rejects
}