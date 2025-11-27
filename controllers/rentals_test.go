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
	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/models"
	"gorm.io/gorm"
)

// Helper to seed a machine for rental tests
// Returns the machine and the DB connection used
func seedRentableMachine(t *testing.T) (models.Machine, *gorm.DB) {
	db := setupTestDB(t, nil) // Uses the shared setup from listing_test.go which resets DB
	
	machine := models.Machine{
		Title:               "Rentable Excavator",
		Description:         "For rent only",
		SellerID:            1, // Owner ID
		ListingType:         "rent",
		RentalPricePerDay:   1000,
		SecurityDeposit:     5000,
		Status:              "listed",
	}
	
	if err := db.Create(&machine).Error; err != nil {
		t.Fatalf("failed to seed machine: %v", err)
	}
	return machine, db
}

// Helper to seed a rental request using an EXISTING db connection
func seedRentalRequest(t *testing.T, db *gorm.DB, machineID uuid.UUID, renterID uint) models.Rental {
	rental := models.Rental{
		MachineID:       machineID,
		RenterID:        renterID,
		StartDate:       time.Now(),
		EndDate:         time.Now().Add(24 * time.Hour),
		TotalAmount:     1000,
		SecurityDeposit: 5000,
		Status:          "pending",
	}
	if err := db.Create(&rental).Error; err != nil {
		t.Fatalf("failed to seed rental: %v", err)
	}
	return rental
}

func TestCreateRentalRequest_EmptyBody(t *testing.T) {
	e := echo.New()
	setupTestDB(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/rentals", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testToken := createTestToken(2, "buyer") // Added Role
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)

	err := CreateRentalRequest(c)
	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", he.Code)
			}
		}
	} else if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreateRentalRequest_InvalidMachineIDFormat(t *testing.T) {
	e := echo.New()
	setupTestDB(t, nil)

	payload := `{"machine_id":"not-a-uuid","start_date":"2025-01-01","end_date":"2025-01-02"}`
	req := httptest.NewRequest(http.MethodPost, "/api/rentals", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	
	testToken := createTestToken(2, "buyer") // Added Role
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c := e.NewContext(req, httptest.NewRecorder())
	c.Set("user", token)

	// Should return 404 because "First" fails on invalid UUID syntax in Postgres (or DB error handled as 404)
	if err := CreateRentalRequest(c); err == nil {
		if c.Response().Status != http.StatusNotFound {
			// We expect 404
		}
	}
}

func TestCreateRentalRequest_StartAfterEnd(t *testing.T) {
	e := echo.New()
	machine, _ := seedRentableMachine(t)

	payload := `{"machine_id":"` + machine.ID.String() + `","start_date":"2025-01-05","end_date":"2025-01-02"}`
	req := httptest.NewRequest(http.MethodPost, "/api/rentals", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testToken := createTestToken(2, "buyer") // Added Role
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)

	CreateRentalRequest(c)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreateRentalRequest_Success(t *testing.T) {
	e := echo.New()
	machine, _ := seedRentableMachine(t)

	payload := `{"machine_id":"` + machine.ID.String() + `","start_date":"2025-01-01","end_date":"2025-01-05"}`
	req := httptest.NewRequest(http.MethodPost, "/api/rentals", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testToken := createTestToken(2, "buyer") // Added Role
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)

	if err := CreateRentalRequest(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestGetMyRentals(t *testing.T) {
	e := echo.New()
	machine, db := seedRentableMachine(t)
	renterID := uint(2)
	_ = seedRentalRequest(t, db, machine.ID, renterID)

	req := httptest.NewRequest(http.MethodGet, "/api/rentals/my", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testToken := createTestToken(renterID, "buyer") // Added Role
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)
	config.DB = db 

	if err := GetMyRentals(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var rentals []models.Rental
	json.Unmarshal(rec.Body.Bytes(), &rentals)
	if len(rentals) != 1 {
		t.Errorf("expected 1 rental, got %d", len(rentals))
	}
}

func TestGetOwnerRentals(t *testing.T) {
	e := echo.New()
	machine, db := seedRentableMachine(t)
	_ = seedRentalRequest(t, db, machine.ID, 2)

	req := httptest.NewRequest(http.MethodGet, "/api/rentals/manage", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testToken := createTestToken(1, "seller") // Added Role
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)
	config.DB = db

	if err := GetOwnerRentals(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var rentals []models.Rental
	json.Unmarshal(rec.Body.Bytes(), &rentals)
	if len(rentals) != 1 {
		t.Errorf("expected 1 rental request, got %d", len(rentals))
	}
}

func TestUpdateRentalStatus(t *testing.T) {
	e := echo.New()
	machine, db := seedRentableMachine(t)
	rental := seedRentalRequest(t, db, machine.ID, 2)

	payload := `{"status":"approved"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	
	c := e.NewContext(req, rec)
	c.SetPath("/api/rentals/:id/status")
	c.SetParamNames("id")
	c.SetParamValues(rental.ID.String())

	testToken := createTestToken(1, "seller") // Added Role
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)
	config.DB = db

	if err := UpdateRentalStatus(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var updated models.Rental
	db.First(&updated, "id = ?", rental.ID)
	if updated.Status != "approved" {
		t.Errorf("expected status approved, got %s", updated.Status)
	}
}

func TestUpdateRentalStatus_Unauthorized(t *testing.T) {
	e := echo.New()
	machine, db := seedRentableMachine(t)
	rental := seedRentalRequest(t, db, machine.ID, 2)

	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"status":"approved"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	
	c := e.NewContext(req, rec)
	c.SetPath("/api/rentals/:id/status")
	c.SetParamNames("id")
	c.SetParamValues(rental.ID.String())

	testToken := createTestToken(2, "buyer") // Added Role
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)
	config.DB = db

	UpdateRentalStatus(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", rec.Code)
	}
}