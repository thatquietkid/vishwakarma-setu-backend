package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	// "strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/vishwakarma-setu-backend/config"
	"github.com/vishwakarma-setu-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	_ = godotenv.Load("../.env")
}

// Helper to generate JWT with Role
func createTestToken(userID uint, role string) string {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"role":    role, // Add Role to claim
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key"
	}
	t, _ := token.SignedString([]byte(secret))
	return t
}

// Helper DB setup
func setupTestDB(t *testing.T, seed []models.Machine) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Fatal("❌ DATABASE_DSN environment variable is not set.")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	_ = db.Migrator().DropTable(&models.Rental{})
	_ = db.Migrator().DropTable(&models.Machine{})
	if err := db.AutoMigrate(&models.Machine{}, &models.Rental{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	if len(seed) > 0 {
		if err := db.Create(&seed).Error; err != nil {
			t.Fatalf("failed to seed: %v", err)
		}
	}
	config.DB = db
	return db
}

// Helper for GET requests
func doRequest(t *testing.T, e *echo.Echo, rawQuery string) (*httptest.ResponseRecorder, error) {
	t.Helper()
	if !strings.HasPrefix(rawQuery, "/") {
		rawQuery = "/" + rawQuery
	}
	req := httptest.NewRequest(http.MethodGet, rawQuery, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := GetAllListings(c); err != nil {
		return rec, err
	}
	return rec, nil
}

// --- TESTS START HERE ---

func TestCreateListing(t *testing.T) {
	e := echo.New()
	setupTestDB(t, nil)

	machineData := `{"title":"New CNC","description":"Brand new","price_for_sale":100000,"listing_type":"sale"}`
	req := httptest.NewRequest(http.MethodPost, "/api/machines", strings.NewReader(machineData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	
	testToken := createTestToken(1, "seller") // Valid role
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+testToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)

	if err := CreateListing(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestCreateListing_Forbidden(t *testing.T) {
	e := echo.New()
	setupTestDB(t, nil)

	machineData := `{"title":"New CNC","description":"Brand new"}`
	req := httptest.NewRequest(http.MethodPost, "/api/machines", strings.NewReader(machineData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	
	// Buyer cannot create listings
	testToken := createTestToken(1, "buyer") 
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+testToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	token, _ := jwt.ParseWithClaims(testToken, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	c.Set("user", token)

	CreateListing(c)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for buyer, got %d", rec.Code)
	}
}

func TestGetListingByID(t *testing.T) {
	e := echo.New()
	seed := []models.Machine{{Title: "Target Machine", SellerID: 1}}
	db := setupTestDB(t, seed)

	var m models.Machine
	db.First(&m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/machines/:id")
	c.SetParamNames("id")
	c.SetParamValues(m.ID.String())

	if err := GetListingByID(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	
	reqNF := httptest.NewRequest(http.MethodGet, "/", nil)
	recNF := httptest.NewRecorder()
	cNF := e.NewContext(reqNF, recNF)
	cNF.SetPath("/api/machines/:id")
	cNF.SetParamNames("id")
	cNF.SetParamValues("00000000-0000-0000-0000-000000000000")

	GetListingByID(cNF)
	if recNF.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", recNF.Code)
	}
}

func TestUpdateListing(t *testing.T) {
	e := echo.New()
	seed := []models.Machine{{Title: "Old Title", SellerID: 1, PriceForSale: 500}}
	db := setupTestDB(t, seed)

	var m models.Machine
	db.First(&m)

	setupCtx := func(body string, userID uint, role string, machineID string) (echo.Context, *httptest.ResponseRecorder) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/machines/:id")
		c.SetParamNames("id")
		c.SetParamValues(machineID)
		
		tokenStr := createTestToken(userID, role)
		token, _ := jwt.ParseWithClaims(tokenStr, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		c.Set("user", token)
		return c, rec
	}

	// Case 1: Success (Owner)
	c1, rec1 := setupCtx(`{"title":"Updated Title"}`, 1, "seller", m.ID.String())
	UpdateListing(c1)
	if rec1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec1.Code)
	}

	// Case 2: Success (Admin Override)
	cAdmin, recAdmin := setupCtx(`{"title":"Admin Edit"}`, 999, "admin", m.ID.String())
	UpdateListing(cAdmin)
	if recAdmin.Code != http.StatusOK {
		t.Errorf("expected 200 for admin, got %d", recAdmin.Code)
	}

	// Case 3: Unauthorized (Wrong User & Not Admin)
	c3, rec3 := setupCtx(`{}`, 999, "seller", m.ID.String())
	UpdateListing(c3)
	if rec3.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec3.Code)
	}
}

func TestDeleteListing(t *testing.T) {
	e := echo.New()
	seed := []models.Machine{{Title: "To Delete", SellerID: 1}}
	db := setupTestDB(t, seed)
	var m models.Machine
	db.First(&m)

	setupCtx := func(userID uint, role string, machineID string) (echo.Context, *httptest.ResponseRecorder) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/machines/:id")
		c.SetParamNames("id")
		c.SetParamValues(machineID)
		tokenStr := createTestToken(userID, role)
		token, _ := jwt.ParseWithClaims(tokenStr, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		c.Set("user", token)
		return c, rec
	}

	// Case 1: Unauthorized
	c2, rec2 := setupCtx(999, "seller", m.ID.String())
	DeleteListing(c2)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec2.Code)
	}

	// Case 2: Success (Owner)
	c3, rec3 := setupCtx(1, "seller", m.ID.String())
	DeleteListing(c3)
	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec3.Code)
	}
}

func TestGetAllListings_Comprehensive(t *testing.T) {
	e := echo.New()
	now := time.Now()
	seed := []models.Machine{
		{Title: "Alpha", Category: "A", ListingType: "sale", PriceForSale: 100, SellerID: 1, CreatedAt: now.Add(-10 * time.Hour)},
		{Title: "Beta", Category: "B", ListingType: "rent", PriceForSale: 200, SellerID: 1, CreatedAt: now.Add(-5 * time.Hour)},
		{Title: "Gamma", Category: "A", ListingType: "both", PriceForSale: 300, SellerID: 1, CreatedAt: now.Add(-1 * time.Hour)},
	}
	setupTestDB(t, seed)

	tests := []struct {
		name          string
		query         string
		expectedTotal int
	}{
		{name: "Search Lower", query: "/?q=alpha", expectedTotal: 1},
		{name: "Search Upper", query: "/?q=BETA", expectedTotal: 1},
		{name: "Filter Sale", query: "/?type=sale", expectedTotal: 2},
		{name: "Filter Rent", query: "/?type=rent", expectedTotal: 2},
		{name: "Filter Category", query: "/?category=A", expectedTotal: 2},
		{name: "Sort Price Asc", query: "/?sort=price_asc", expectedTotal: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec, err := doRequest(t, e, tc.query)
			if err != nil {
				t.Fatalf("handler error: %v", err)
			}
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 OK")
			}
			var resp map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &resp)
			if total, ok := resp["total"].(float64); ok {
				if int(total) != tc.expectedTotal {
					t.Errorf("expected total %d, got %d", tc.expectedTotal, int(total))
				}
			}
		})
	}
}