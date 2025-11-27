package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/vishwakarma-setu-backend/models"
	"gorm.io/gorm"
)

// Helper to seed maintenance records
func seedMaintenanceRecord(t *testing.T, db *gorm.DB, machineID uuid.UUID) models.MaintenanceRecord {
	record := models.MaintenanceRecord{
		MachineID:   machineID,
		ServiceDate: time.Now(),
		Type:        "Routine",
		Description: "Oil change and filter replacement",
		Cost:        5000.00,
		Technician:  "Rajesh Kumar",
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("failed to seed maintenance record: %v", err)
	}
	return record
}

func TestAddMaintenanceRecord_Success(t *testing.T) {
	e := echo.New()
	// 1. Seed machine (Owner ID 1)
	// Note: seedRentableMachine calls setupTestDB which resets the DB
	machine, db := seedRentableMachine(t)

	// 2. Ensure Maintenance table exists (in case setupTestDB didn't include it)
	if err := db.AutoMigrate(&models.MaintenanceRecord{}); err != nil {
		t.Fatalf("failed to migrate maintenance table: %v", err)
	}

	// 3. Prepare Payload
	payload := `{
		"machine_id": "` + machine.ID.String() + `",
		"service_date": "2025-01-15",
		"type": "Repair",
		"description": "Fixed hydraulic leak",
		"cost": 12000.50,
		"technician": "Service Pro Inc."
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/maintenance", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 4. Mock Auth (Owner ID 1)
	testToken := createTestToken(1, "seller")
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)

	// 5. Execute Handler
	if err := AddMaintenanceRecord(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	// 6. Assertions
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp models.MaintenanceRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response json: %v", err)
	}

	if resp.Type != "Repair" {
		t.Errorf("expected type 'Repair', got '%s'", resp.Type)
	}
	if resp.MachineID != machine.ID {
		t.Errorf("expected machine ID match")
	}
}

func TestAddMaintenanceRecord_Unauthorized_NotOwner(t *testing.T) {
	e := echo.New()
	machine, db := seedRentableMachine(t) // Owner is ID 1
	db.AutoMigrate(&models.MaintenanceRecord{})

	payload := `{
		"machine_id": "` + machine.ID.String() + `",
		"service_date": "2025-01-15",
		"type": "Repair",
		"description": "Unauthorized fix",
		"cost": 100.00,
		"technician": "Self"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/maintenance", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock Auth (User ID 2 - NOT the owner)
	testToken := createTestToken(2, "buyer")
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)

	AddMaintenanceRecord(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden, got %d", rec.Code)
	}
}

func TestAddMaintenanceRecord_InvalidMachineID(t *testing.T) {
	e := echo.New()
	_, db := seedRentableMachine(t)
	db.AutoMigrate(&models.MaintenanceRecord{})

	// Non-existent machine ID
	fakeID := "00000000-0000-0000-0000-000000000000"
	payload := `{
		"machine_id": "` + fakeID + `",
		"service_date": "2025-01-15",
		"type": "Repair",
		"description": "Ghost machine",
		"cost": 0,
		"technician": "Ghostbuster"
	}`

	req := httptest.NewRequest(http.MethodPost, "/api/maintenance", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Auth as valid user
	testToken := createTestToken(1, "seller")
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)

	AddMaintenanceRecord(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 Not Found, got %d", rec.Code)
	}
}

func TestGetMaintenanceHistory_Success(t *testing.T) {
	e := echo.New()
	machine, db := seedRentableMachine(t)
	db.AutoMigrate(&models.MaintenanceRecord{})

	// Seed 2 records
	seedMaintenanceRecord(t, db, machine.ID)
	seedMaintenanceRecord(t, db, machine.ID)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/machines/:machine_id/maintenance")
	c.SetParamNames("machine_id")
	c.SetParamValues(machine.ID.String())

	if err := GetMaintenanceHistory(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var records []models.MaintenanceRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &records); err != nil {
		t.Fatalf("invalid response json: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
	// Verify sorting (default desc) is handled by DB, but we just check count here
}

func TestGetMaintenanceHistory_Empty(t *testing.T) {
	e := echo.New()
	machine, db := seedRentableMachine(t)
	db.AutoMigrate(&models.MaintenanceRecord{})

	// No records seeded

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/machines/:machine_id/maintenance")
	c.SetParamNames("machine_id")
	c.SetParamValues(machine.ID.String())

	if err := GetMaintenanceHistory(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var records []models.MaintenanceRecord
	if err := json.Unmarshal(rec.Body.Bytes(), &records); err != nil {
		t.Fatalf("invalid response json: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}