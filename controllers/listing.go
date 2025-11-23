package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/models"
)

// Helper function to extract user ID from token (unchanged)
func getUserIDFromToken(c echo.Context) (uint, error) {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*jwt.MapClaims)

	if userIDFloat, ok := (*claims)["user_id"].(float64); ok {
		return uint(userIDFloat), nil
	}

	return 0, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims: user_id missing or invalid")
}

// CreateListing (unchanged)
func CreateListing(c echo.Context) error {
	var machine models.Machine
	if err := c.Bind(&machine); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input data"})
	}

	sellerID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	machine.SellerID = sellerID

	if err := config.DB.Create(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create listing: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, machine)
}

// GetAllListings - GET /api/machines (Search Implementation)
func GetAllListings(c echo.Context) error {
	var machines []models.Machine
	query := config.DB.Model(&models.Machine{})

	// 1. Keyword Search (Title & Description)
	if q := c.QueryParam("q"); q != "" {
		searchTerm := "%" + strings.ToLower(q) + "%"
		// UPDATED: Use LOWER() LIKE LOWER() for SQLite compatibility in tests
		query = query.Where("LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", searchTerm, searchTerm)
	}

	// 2. Exact Filters
	if category := c.QueryParam("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if manufacturer := c.QueryParam("manufacturer"); manufacturer != "" {
		query = query.Where("manufacturer = ?", manufacturer)
	}
	if location := c.QueryParam("location"); location != "" {
		// UPDATED: Use LOWER() LIKE LOWER() for SQLite compatibility in tests
		query = query.Where("LOWER(location) LIKE LOWER(?)", "%"+location+"%")
	}

	// 3. Listing Type Filter (sale, rent)
	if listingType := c.QueryParam("type"); listingType != "" {
		if listingType == "sale" {
			query = query.Where("listing_type IN (?, ?)", "sale", "both")
		} else if listingType == "rent" {
			query = query.Where("listing_type IN (?, ?)", "rent", "both")
		}
	}

	// 4. Price Range Filter
	if minPrice := c.QueryParam("min_price"); minPrice != "" {
		query = query.Where("price_for_sale >= ?", minPrice)
	}
	if maxPrice := c.QueryParam("max_price"); maxPrice != "" {
		query = query.Where("price_for_sale <= ?", maxPrice)
	}

	// 5. Sorting
	sortParam := c.QueryParam("sort")
	switch sortParam {
	case "price_asc":
		query = query.Order("price_for_sale asc")
	case "price_desc":
		query = query.Order("price_for_sale desc")
	case "oldest":
		query = query.Order("created_at asc")
	default:
		query = query.Order("created_at desc") // Default: Newest first
	}

	// 6. Pagination
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page <= 0 { page = 1 }
	if limit <= 0 { limit = 10 }
	offset := (page - 1) * limit

	// Execute Query
	var total int64
	query.Count(&total) 
	
	if err := query.Offset(offset).Limit(limit).Find(&machines).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not fetch listings"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":       machines,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetListingByID (unchanged)
func GetListingByID(c echo.Context) error {
	id := c.Param("id")
	var machine models.Machine
	if err := config.DB.First(&machine, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}
	return c.JSON(http.StatusOK, machine)
}

// UpdateListing (unchanged)
func UpdateListing(c echo.Context) error {
	id := c.Param("id")
	var machine models.Machine

	if err := config.DB.First(&machine, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}

	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	if machine.SellerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this listing"})
	}

	var updateData models.Machine
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// Explicitly update fields
	machine.Title = updateData.Title
	machine.Description = updateData.Description
	machine.Category = updateData.Category       
	machine.Location = updateData.Location       
	machine.PriceForSale = updateData.PriceForSale
	machine.RentalPricePerMonth = updateData.RentalPricePerMonth
	machine.RentalPricePerDay = updateData.RentalPricePerDay
	machine.SecurityDeposit = updateData.SecurityDeposit
	machine.Specs = updateData.Specs
	machine.Status = updateData.Status
	machine.ListingType = updateData.ListingType

	config.DB.Save(&machine)
	return c.JSON(http.StatusOK, machine)
}

// DeleteListing (unchanged)
func DeleteListing(c echo.Context) error {
	id := c.Param("id")
	var machine models.Machine

	if err := config.DB.First(&machine, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}

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