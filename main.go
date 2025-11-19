package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"os"
	_ "github.com/joho/godotenv/autoload"

	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/routes"
)

func main() {
	e := echo.New()

	// Connect to the database
	// This will also run migrations if not in production
	config.ConnectDatabase()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{os.Getenv("FRONTEND_URL"),"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
	}))

	// Set up routes
	routes.RegisterRoutes(e)

	// Start the server
	// Listen on port 1323
	port := os.Getenv("PORT")
	if port == "" {
		port = "1324"
	}
	e.Logger.Fatal(e.Start(":" + port))
}