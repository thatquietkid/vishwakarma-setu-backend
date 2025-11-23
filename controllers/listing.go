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

// Helper function to extract user ID from token
func getUserIDFromToken(c echo.Context) (uint, error) {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*jwt.MapClaims)

	if userIDFloat, ok := (*claims)["user_id"].(float64); ok {
		return uint(userIDFloat), nil
	}

	return 0, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token claims: user_id missing or invalid")
}

// CreateListing godoc
// @Summary Create a new machine listing
// @Description Register a new machine for sale or rent. Requires a valid Seller JWT token.
// @Tags Machines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param machine body models.Machine true "Machine Details"
// @Success 201 {object} models.Machine
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /machines [post]
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

// GetAllListings godoc
// @Summary Get all machine listings
// @Description Retrieve a list of machines with optional filtering, sorting, and pagination.
// @Tags Machines
// @Produce json
// @Param q query string false "Search query (Title/Description)"
// @Param category query string false "Filter by Category"
// @Param manufacturer query string false "Filter by Manufacturer"
// @Param location query string false "Filter by Location"
// @Param type query string false "Filter by Listing Type (sale, rent)"
// @Param min_price query number false "Minimum Price (Sale)"
// @Param max_price query number false "Maximum Price (Sale)"
// @Param sort query string false "Sort order (price_asc, price_desc, oldest)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} map[string]interface{} "Returns data array and pagination meta"
// @Failure 500 {object} map[string]string
// @Router /machines [get]
func GetAllListings(c echo.Context) error {
	var machines []models.Machine
	query := config.DB.Model(&models.Machine{})

	// 1. Keyword Search
	if q := c.QueryParam("q"); q != "" {
		searchTerm := "%" + strings.ToLower(q) + "%"
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
		query = query.Where("LOWER(location) LIKE LOWER(?)", "%"+location+"%")
	}

	// 3. Listing Type Filter
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
		query = query.Order("created_at desc")
	}

	// 6. Pagination
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page <= 0 { page = 1 }
	if limit <= 0 { limit = 10 }
	offset := (page - 1) * limit

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

// GetListingByID godoc
// @Summary Get a machine by ID
// @Description Retrieve full details of a specific machine listing.
// @Tags Machines
// @Produce json
// @Param id path string true "Machine ID (UUID)"
// @Success 200 {object} models.Machine
// @Failure 404 {object} map[string]string "Machine not found"
// @Router /machines/{id} [get]
func GetListingByID(c echo.Context) error {
	id := c.Param("id")
	var machine models.Machine
	if err := config.DB.First(&machine, "id = ?", id).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}
	return c.JSON(http.StatusOK, machine)
}

// UpdateListing godoc
// @Summary Update a listing
// @Description Update details of an existing machine listing. Only the owner can perform this.
// @Tags Machines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Machine ID"
// @Param machine body models.Machine true "Updated Machine Data"
// @Success 200 {object} models.Machine
// @Failure 403 {object} map[string]string "Not authorized"
// @Failure 404 {object} map[string]string "Machine not found"
// @Router /machines/{id} [put]
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

	// Strict Ownership Check
	if machine.SellerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this listing"})
	}

	var updateData models.Machine
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

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

// DeleteListing godoc
// @Summary Delete a listing
// @Description Permanently remove a machine listing. Only the owner can perform this.
// @Tags Machines
// @Produce json
// @Security BearerAuth
// @Param id path string true "Machine ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 403 {object} map[string]string "Not authorized"
// @Failure 404 {object} map[string]string "Machine not found"
// @Router /machines/{id} [delete]
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

	// Strict Ownership Check
	if machine.SellerID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to delete this listing"})
	}

	if err := config.DB.Delete(&machine).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete listing"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Listing deleted successfully"})
}