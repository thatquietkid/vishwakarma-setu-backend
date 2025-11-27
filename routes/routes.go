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

	// 1. Static File Serving
	e.Static("/uploads", "./uploads")

	// API Group
	api := e.Group("/api")

	// Root & Utility Routes
	api.GET("/", controllers.Index)
	api.GET("/health", controllers.HealthCheck)

	// Error handling routes
	api.GET("/not-found", controllers.NotFound)
	api.GET("/internal-server-error", controllers.InternalServerError)
	api.GET("/bad-request", controllers.BadRequest)
	api.GET("/unauthorized", controllers.Unauthorized)
	api.GET("/forbidden", controllers.Forbidden)

	// Public Machine Routes
	api.GET("/machines", controllers.GetAllListings)
	api.GET("/machines/:id", controllers.GetListingByID)

	// Public Inspection Route (Buyers need to see the report)
	api.GET("/machines/:machine_id/inspection", controllers.GetMachineInspection)

	// Public Maintenance Route
	api.GET("/machines/:machine_id/maintenance", controllers.GetMaintenanceHistory)

	// Protected Routes (Auth Required)
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware())

	// Utility
	protected.POST("/upload", controllers.UploadImage)

	// Machine Management
	protected.POST("/machines", controllers.CreateListing)
	protected.PUT("/machines/:id", controllers.UpdateListing)
	protected.DELETE("/machines/:id", controllers.DeleteListing)

	// Rental Management
	protected.POST("/rentals", controllers.CreateRentalRequest)
	protected.GET("/rentals/my", controllers.GetMyRentals)
	protected.GET("/rentals/manage", controllers.GetOwnerRentals)
	protected.PUT("/rentals/:id/status", controllers.UpdateRentalStatus)

	// Inspection Management
	protected.POST("/inspections", controllers.CreateInspectionReport) // Inspector submits report

	// Protected Maintenance Route
	protected.POST("/maintenance", controllers.AddMaintenanceRecord)
}
