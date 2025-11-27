package controllers

import (
	"math"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/models"
)

// RentalRequest represents the payload to create a rental
type RentalRequest struct {
	MachineID string `json:"machine_id" example:"uuid-string"`
	StartDate string `json:"start_date" example:"2025-01-01"` // Format: YYYY-MM-DD
	EndDate   string `json:"end_date" example:"2025-01-05"`   // Format: YYYY-MM-DD
}

// RentalStatusUpdate represents the payload to update status
type RentalStatusUpdate struct {
	Status string `json:"status" example:"approved"` // approved, rejected, completed
}

// CreateRentalRequest godoc
//
//	@Summary		Request to rent a machine
//	@Description	Initiate a rental request. Calculates fees based on duration. Requires Renter Auth.
//	@Tags			Rentals
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		RentalRequest	true	"Rental Details"
//	@Success		201		{object}	models.Rental
//	@Failure		400		{object}	map[string]string	"Invalid input or dates"
//	@Failure		404		{object}	map[string]string	"Machine not found"
//	@Router			/rentals [post]
func CreateRentalRequest(c echo.Context) error {
	var req RentalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if req.MachineID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Machine ID is required"})
	}

	// FIX: Use shared helper from listing.go
	user, err := getUserClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	var machine models.Machine
	if err := config.DB.First(&machine, "id = ?", req.MachineID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}

	if machine.ListingType == "sale" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "This machine is not for rent"})
	}

	layout := "2006-01-02"
	start, err := time.Parse(layout, req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid start_date format"})
	}
	end, err := time.Parse(layout, req.EndDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid end_date format"})
	}

	if end.Before(start) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "End date cannot be before start date"})
	}

	days := math.Ceil(end.Sub(start).Hours() / 24)
	if days < 1 {
		days = 1
	}

	rentalFee := days * machine.RentalPricePerDay
	platformFee := rentalFee * 0.05 // 5% Commission

	rental := models.Rental{
		MachineID:       machine.ID,
		RenterID:        user.ID, // Use ID from claims
		StartDate:       start,
		EndDate:         end,
		TotalAmount:     rentalFee,
		SecurityDeposit: machine.SecurityDeposit,
		PlatformFee:     platformFee,
		Status:          "pending",
	}

	if err := config.DB.Create(&rental).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create rental request"})
	}

	return c.JSON(http.StatusCreated, rental)
}

// GetMyRentals godoc
//
//	@Summary		Get my rental history
//	@Description	Retrieve all rental requests made by the logged-in user.
//	@Tags			Rentals
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}	models.Rental
//	@Router			/rentals/my [get]
func GetMyRentals(c echo.Context) error {
	// FIX: Use shared helper
	user, err := getUserClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	var rentals []models.Rental
	if err := config.DB.Preload("Machine").Where("renter_id = ?", user.ID).Find(&rentals).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch rentals"})
	}

	return c.JSON(http.StatusOK, rentals)
}

// GetOwnerRentals godoc
//
//	@Summary		Get rental requests for my machines
//	@Description	Retrieve all incoming rental requests for machines owned by the logged-in user.
//	@Tags			Rentals
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}	models.Rental
//	@Router			/rentals/manage [get]
func GetOwnerRentals(c echo.Context) error {
	// FIX: Use shared helper
	user, err := getUserClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	var rentals []models.Rental
	err = config.DB.Preload("Machine").
		Joins("JOIN machines ON machines.id = rentals.machine_id").
		Where("machines.seller_id = ?", user.ID).
		Find(&rentals).Error

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch requests"})
	}

	return c.JSON(http.StatusOK, rentals)
}

// UpdateRentalStatus godoc
//
//	@Summary		Update rental status
//	@Description	Approve, reject, or complete a rental. Only the machine owner can do this.
//	@Tags			Rentals
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"Rental ID"
//	@Param			status	body		RentalStatusUpdate	true	"New Status"
//	@Success		200		{object}	models.Rental
//	@Failure		403		{object}	map[string]string	"Not authorized"
//	@Router			/rentals/{id}/status [put]
func UpdateRentalStatus(c echo.Context) error {
	rentalID := c.Param("id")

	var req RentalStatusUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// FIX: Use shared helper
	user, err := getUserClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	var rental models.Rental
	if err := config.DB.Preload("Machine").First(&rental, "id = ?", rentalID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Rental not found"})
	}

	if rental.Machine.SellerID != user.ID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not the owner of this machine"})
	}

	rental.Status = req.Status
	config.DB.Save(&rental)

	return c.JSON(http.StatusOK, rental)
}
