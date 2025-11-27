package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/models"
	"gorm.io/datatypes"
)

// InspectionRequest payload
type InspectionRequest struct {
	MachineID  string                 `json:"machine_id"`
	ReportType string                 `json:"report_type"` // listing, check_out, check_in
	Verdict    string                 `json:"verdict"`
	Summary    string                 `json:"summary"`
	ReportData map[string]interface{} `json:"report_data"` // Flexible Key-Value pairs
	MediaURLs  []string               `json:"media_urls"`  // Array of image URLs
}

// CreateInspectionReport godoc
//
//	@Summary		Submit an inspection report
//	@Description	Submit a verification report with media URLs. Only Inspectors/Admins should ideally do this.
//	@Tags			Inspection
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			report	body		InspectionRequest	true	"Inspection Data"
//	@Success		201		{object}	models.InspectionReport
//	@Failure		400		{object}	map[string]string	"Invalid Input"
//	@Router			/inspections [post]
func CreateInspectionReport(c echo.Context) error {
	var req InspectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	// FIX: Use the new shared helper from listing.go
	user, err := getUserClaims(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	// Optional: Enforce Role (e.g., only inspectors or admins)
	// if user.Role != "inspector" && user.Role != "admin" {
	//     return c.JSON(http.StatusForbidden, map[string]string{"error": "Only inspectors can submit reports"})
	// }

	// 1. Verify Machine Exists
	var machine models.Machine
	if err := config.DB.First(&machine, "id = ?", req.MachineID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Machine not found"})
	}

	// 2. Serialize JSON fields
	reportDataJSON, _ := json.Marshal(req.ReportData)
	mediaURLsJSON, _ := json.Marshal(req.MediaURLs)

	report := models.InspectionReport{
		MachineID:      machine.ID,
		InspectorID:    user.ID, // Use ID from claims
		ReportType:     req.ReportType,
		InspectionDate: time.Now(),
		Verdict:        req.Verdict,
		Summary:        req.Summary,
		ReportData:     datatypes.JSON(reportDataJSON),
		MediaURLs:      datatypes.JSON(mediaURLsJSON),
	}

	if err := config.DB.Create(&report).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save report"})
	}

	// 3. Optional: Update Machine Status to 'verified' if it was pending
	if machine.Status == "pending_inspection" && req.Verdict != "Fail" {
		config.DB.Model(&machine).Update("status", "verified")
	}

	return c.JSON(http.StatusCreated, report)
}

// GetMachineInspection godoc
//
//	@Summary		Get inspection report for a machine
//	@Description	Retrieve the latest inspection report for a specific machine.
//	@Tags			Inspection
//	@Produce		json
//	@Param			machine_id	path		string	true	"Machine UUID"
//	@Success		200			{object}	models.InspectionReport
//	@Router			/machines/{machine_id}/inspection [get]
func GetMachineInspection(c echo.Context) error {
	machineID := c.Param("machine_id")

	var report models.InspectionReport
	// Get the most recent report for this machine
	if err := config.DB.Where("machine_id = ?", machineID).Order("created_at desc").First(&report).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No inspection report found for this machine"})
	}

	return c.JSON(http.StatusOK, report)
}
