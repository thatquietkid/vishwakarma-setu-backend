package controllers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/models"
)

// Helper function to extract user ID from token
func getUserIDFromToken(c echo.Context) (uint, error) {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*jwt.MapClaims)

	// JWT numbers are often unmarshaled as float64
	if userIDFloat, ok := (*claims)["user_id"].(float64); ok {
		return uint(userIDFloat), nil
	}

	return 0, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims: user_id missing or invalid")
}

// CreateListing - POST /api/machines
func CreateListing(c echo.Context) error {
	var machine models.Machine
	if err := c.Bind(&machine); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input data"})
	}

	// Extract User/Seller ID from JWT Token
	sellerID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	// Force the SellerID to match the token owner
	machine.SellerID = sellerID

	if err := config.DB.Create(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create listing: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, machine)
}

// GetAllListings - GET /api/machines
func GetAllListings(c echo.Context) error {
	var machines []models.Machine
	
	listingType := c.QueryParam("type") // "sale" or "rent"
	query := config.DB

	if listingType != "" {
		query = query.Where("listing_type = ? OR listing_type = 'both'", listingType)
	}

	if err := query.Find(&machines).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not fetch listings"})
	}

	return c.JSON(http.StatusOK, machines)
}

// GetListingByID - GET /api/machines/:id
func GetListingByID(c echo.Context) error {
	id := c.Param("id")
	var machine models.Machine

	if err := config.DB.First(&machine, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}

	return c.JSON(http.StatusOK, machine)
}

// UpdateListing - PUT /api/machines/:id
func UpdateListing(c echo.Context) error {
	id := c.Param("id")
	var machine models.Machine

	// 1. Find the machine
	if err := config.DB.First(&machine, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}

	// 2. Verify Ownership
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	if machine.SellerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this listing"})
	}

	// 3. Update data
	var updateData models.Machine
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Update fields
	machine.Title = updateData.Title
	machine.Description = updateData.Description
	machine.PriceForSale = updateData.PriceForSale
	machine.RentalPricePerMonth = updateData.RentalPricePerMonth
	machine.Specs = updateData.Specs
	machine.Status = updateData.Status
	machine.ListingType = updateData.ListingType

	config.DB.Save(&machine)
	return c.JSON(http.StatusOK, machine)
}

// DeleteListing - DELETE /api/machines/:id
func DeleteListing(c echo.Context) error {
	id := c.Param("id")
	var machine models.Machine

	if err := config.DB.First(&machine, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}

	// Verify Ownership
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	if machine.SellerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to delete this listing"})
	}

	if err := config.DB.Delete(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete listing"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Listing deleted successfully"})
}