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

// Helper to generate JWT
func createTestToken(userID uint) string {
	claims := jwt.MapClaims{
		"user_id": float64(userID),
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
	err = db.Migrator().DropTable(&models.Machine{})
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}
	if err := db.AutoMigrate(&models.Machine{}); err != nil {
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

	// Case 1: Success
	machineData := `{"title":"New CNC","description":"Brand new","price_for_sale":100000,"listing_type":"sale"}`
	req := httptest.NewRequest(http.MethodPost, "/api/machines", strings.NewReader(machineData))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	
	testToken := createTestToken(1)
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

	// Case 2: Invalid JSON (Bad Request)
	reqBad := httptest.NewRequest(http.MethodPost, "/api/machines", strings.NewReader(`{"title": "Broken JSON`)) // Missing closing brace
	reqBad.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recBad := httptest.NewRecorder()
	cBad := e.NewContext(reqBad, recBad)
	
	// We skip auth middleware here for unit testing simplicity, as Bind happens before logic usually
	// But since controller extracts ID first, we must provide context
	cBad.Set("user", token) 

	if err := CreateListing(cBad); err == nil {
		// It usually returns an error struct that Echo handles, or we return c.JSON(400)
		// In our controller we return c.JSON(400), so err might be nil but code should be 400
		if recBad.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for invalid JSON, got %d", recBad.Code)
		}
	}
}

func TestGetListingByID(t *testing.T) {
	e := echo.New()
	seed := []models.Machine{{Title: "Target Machine", SellerID: 1}}
	db := setupTestDB(t, seed)

	var m models.Machine
	db.First(&m)

	// Case 1: Success
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

	// Case 2: Not Found
	reqNF := httptest.NewRequest(http.MethodGet, "/", nil)
	recNF := httptest.NewRecorder()
	cNF := e.NewContext(reqNF, recNF)
	cNF.SetPath("/api/machines/:id")
	cNF.SetParamNames("id")
	cNF.SetParamValues("00000000-0000-0000-0000-000000000000") // Valid UUID format, but non-existent

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

	// Setup Context Helper
	setupCtx := func(body string, userID uint, machineID string) (echo.Context, *httptest.ResponseRecorder) {
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/machines/:id")
		c.SetParamNames("id")
		c.SetParamValues(machineID)
		
		tokenStr := createTestToken(userID)
		token, _ := jwt.ParseWithClaims(tokenStr, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		c.Set("user", token)
		return c, rec
	}

	// Case 1: Success
	c1, rec1 := setupCtx(`{"title":"Updated Title"}`, 1, m.ID.String())
	UpdateListing(c1)
	if rec1.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec1.Code)
	}

	// Case 2: Not Found
	c2, rec2 := setupCtx(`{}`, 1, "00000000-0000-0000-0000-000000000000")
	UpdateListing(c2)
	if rec2.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec2.Code)
	}

	// Case 3: Unauthorized (Wrong User)
	c3, rec3 := setupCtx(`{}`, 999, m.ID.String())
	UpdateListing(c3)
	if rec3.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec3.Code)
	}

	// Case 4: Invalid JSON (Binding Error)
	c4, rec4 := setupCtx(`{"title": "Brok`, 1, m.ID.String())
	UpdateListing(c4)
	if rec4.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec4.Code)
	}
}

func TestDeleteListing(t *testing.T) {
	e := echo.New()
	seed := []models.Machine{{Title: "To Delete", SellerID: 1}}
	db := setupTestDB(t, seed)
	var m models.Machine
	db.First(&m)

	setupCtx := func(userID uint, machineID string) (echo.Context, *httptest.ResponseRecorder) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/machines/:id")
		c.SetParamNames("id")
		c.SetParamValues(machineID)
		tokenStr := createTestToken(userID)
		token, _ := jwt.ParseWithClaims(tokenStr, new(jwt.MapClaims), func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		c.Set("user", token)
		return c, rec
	}

	// Case 1: Not Found
	c1, rec1 := setupCtx(1, "00000000-0000-0000-0000-000000000000")
	DeleteListing(c1)
	if rec1.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec1.Code)
	}

	// Case 2: Unauthorized
	c2, rec2 := setupCtx(999, m.ID.String())
	DeleteListing(c2)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec2.Code)
	}

	// Case 3: Success
	c3, rec3 := setupCtx(1, m.ID.String())
	DeleteListing(c3)
	if rec3.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec3.Code)
	}
}

func TestGetAllListings_Comprehensive(t *testing.T) {
	e := echo.New()
	now := time.Now()
	
	// Create diverse dataset
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
		expectedFirst string // Title of the first item (for sort checks)
	}{
		// 1. Search Term (Case Insensitive)
		{name: "Search Lower", query: "/?q=alpha", expectedTotal: 1, expectedFirst: "Alpha"},
		{name: "Search Upper", query: "/?q=BETA", expectedTotal: 1, expectedFirst: "Beta"},
		
		// 2. Filters
		{name: "Filter Sale", query: "/?type=sale", expectedTotal: 2, expectedFirst: ""}, // Includes 'sale' and 'both' (Alpha, Gamma)
		{name: "Filter Rent", query: "/?type=rent", expectedTotal: 2, expectedFirst: ""}, // Includes 'rent' and 'both' (Beta, Gamma)
		{name: "Filter Category", query: "/?category=A", expectedTotal: 2, expectedFirst: ""},

		// 3. Sorting
		{name: "Sort Price Asc", query: "/?sort=price_asc", expectedTotal: 3, expectedFirst: "Alpha"}, // 100
		{name: "Sort Price Desc", query: "/?sort=price_desc", expectedTotal: 3, expectedFirst: "Gamma"}, // 300
		{name: "Sort Oldest", query: "/?sort=oldest", expectedTotal: 3, expectedFirst: "Alpha"}, // Created 10h ago
		{name: "Sort Newest (Default)", query: "/?sort=newest", expectedTotal: 3, expectedFirst: "Gamma"}, // Created 1h ago
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

			// Check Total
			if total, ok := resp["total"].(float64); ok {
				if int(total) != tc.expectedTotal {
					t.Errorf("expected total %d, got %d", tc.expectedTotal, int(total))
				}
			}

			// Check Sort Order (if expectedFirst is set)
			if tc.expectedFirst != "" {
				data := resp["data"].([]interface{})
				if len(data) > 0 {
					first := data[0].(map[string]interface{})
					if first["title"] != tc.expectedFirst {
						t.Errorf("expected first item '%s', got '%s'", tc.expectedFirst, first["title"])
					}
				}
			}
		})
	}
}