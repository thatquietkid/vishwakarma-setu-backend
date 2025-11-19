package middleware

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// JWTMiddlewareConfig returns the Echo JWT middleware configuration
func JWTMiddleware() echo.MiddlewareFunc {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Fallback or panic in production
		secret = "your-secret-key" 
	}

	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwt.MapClaims)
		},
		SigningKey: []byte(secret),
	}

	return echojwt.WithConfig(config)
}