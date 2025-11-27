package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/routes"

	_ "github.com/vishwakarma-setu-backend/docs" // Import generated docs
)

// @title						Vishwakarma Setu API
// @version					1.0
// @description				Backend API for Vishwakarma Setu B2B Marketplace.
// @host						localhost:1324
// @BasePath					/api
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	// Load env
	_ = godotenv.Load()

	e := echo.New()

	// Connect to the database
	config.ConnectDatabase()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{os.Getenv("FRONTEND_URL"), "*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
	}))

	// Set up routes
	routes.RegisterRoutes(e)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "1324"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
