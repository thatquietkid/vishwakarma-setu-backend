package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestHandlers(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name          string
		handler       func(echo.Context) error
		expectedCode  int
		expectedKey   string
		expectedValue string
	}{
		{"Index", Index, http.StatusOK, "message", "Welcome to the Student Dashboard API"},
		{"HealthCheck", HealthCheck, http.StatusOK, "status", "OK"},
		{"NotFound", NotFound, http.StatusNotFound, "error", "Resource not found"},
		{"InternalServerError", InternalServerError, http.StatusInternalServerError, "error", "Internal server error"},
		{"BadRequest", BadRequest, http.StatusBadRequest, "error", "Bad request"},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized, "error", "Unauthorized access"},
		{"Forbidden", Forbidden, http.StatusForbidden, "error", "Forbidden access"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			if err := tc.handler(ctx); err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if rec.Code != tc.expectedCode {
				t.Fatalf("expected status %d, got %d", tc.expectedCode, rec.Code)
			}

			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid json response: %v", err)
			}

			if v, ok := body[tc.expectedKey]; !ok {
				t.Fatalf("response missing key %q", tc.expectedKey)
			} else if v != tc.expectedValue {
				t.Fatalf("expected %q for key %q, got %q", tc.expectedValue, tc.expectedKey, v)
			}
		})
	}
}
