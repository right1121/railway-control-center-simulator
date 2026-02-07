package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/right1121/railway-control-center-simulator/internal/config"
	"github.com/right1121/railway-control-center-simulator/internal/di"
)

func TestSetupRegistersSimulationRoute(t *testing.T) {
	cfg := &config.Config{}
	container := di.NewContainer(cfg)
	mux := setup(http.NewServeMux(), cfg, container)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulation", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatalf("expected simulation route to be registered, got 404")
	}
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 200 or 500, got %d", rec.Code)
	}
}

func TestSetupRegistersSimulationTickRoute(t *testing.T) {
	cfg := &config.Config{}
	container := di.NewContainer(cfg)
	mux := setup(http.NewServeMux(), cfg, container)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/tick", strings.NewReader(`{"deltaMillis":`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
