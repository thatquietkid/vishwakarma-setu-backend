package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/models"
)

type MaintenanceRequest struct {
	MachineID   string  `json:"machine_id"`
	ServiceDate string  `json:"service_date"` // YYYY-MM-DD
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Cost        float64 `json:"cost"`
	Technician  string  `json:"technician"`
	DocumentURL string  `json:"document_url"`
}

// AddMaintenanceRecord godoc
//
//	@Summary		Add a maintenance record
//	@Description	Add a service history log for a machine. Only the owner can do this.
//	@Tags			Maintenance
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			record	body		MaintenanceRequest	true	"Maintenance Data"
//	@Success		201		{object}	models.MaintenanceRecord
//	@Failure		400		{object}	map[string]string	"Invalid input"
//	@Failure		403		{object}	map[string]string	"Not authorized"
//	@Router			/maintenance [post]
func AddMaintenanceRecord(c echo.Context) error {
	var req MaintenanceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	ownerID, err := getUserClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Verify ownership
	var machine models.Machine
	if err := config.DB.First(&machine, "id = ?", req.MachineID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}

	if machine.SellerID != ownerID.ID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not the owner of this machine"})
	}

	date, _ := time.Parse("2006-01-02", req.ServiceDate)

	record := models.MaintenanceRecord{
		MachineID:   machine.ID,
		ServiceDate: date,
		Type:        req.Type,
		Description: req.Description,
		Cost:        req.Cost,
		Technician:  req.Technician,
		DocumentURL: req.DocumentURL,
	}

	if err := config.DB.Create(&record).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save record"})
	}

	return c.JSON(http.StatusCreated, record)
}

// GetMaintenanceHistory godoc
//
//	@Summary		Get maintenance history
//	@Description	Retrieve full service history for a machine.
//	@Tags			Maintenance
//	@Produce		json
//	@Param			machine_id	path	string	true	"Machine ID"
//	@Success		200			{array}	models.MaintenanceRecord
//	@Router			/machines/{machine_id}/maintenance [get]
func GetMaintenanceHistory(c echo.Context) error {
	machineID := c.Param("machine_id")
	var records []models.MaintenanceRecord

	if err := config.DB.Where("machine_id = ?", machineID).Order("service_date desc").Find(&records).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch records"})
	}

	return c.JSON(http.StatusOK, records)
}
