package http

import (
	"encoding/json"
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

func TestSetupRegistersSimulationPermissionRouteAndRejectsBadJSON(t *testing.T) {
	cfg := &config.Config{}
	cfg.Simulation.LineDefinitionPath = "internal/domain/simulation/fixtures/line.json"
	container := di.NewContainer(cfg)
	mux := setup(http.NewServeMux(), cfg, container)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/permission", strings.NewReader(`{"stationId":`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertRouterErrorBody(t, rec.Body.Bytes(), "BAD_JSON", "invalid json")
}

func TestSimulationPermissionReturnsBadRequestOnInvalidStationID(t *testing.T) {
	cfg := &config.Config{}
	cfg.Simulation.LineDefinitionPath = "internal/domain/simulation/fixtures/line.json"
	container := di.NewContainer(cfg)
	mux := setup(http.NewServeMux(), cfg, container)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/permission", strings.NewReader(`{"stationId":" ","allowed":true}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertRouterErrorBody(t, rec.Body.Bytes(), "INVALID_STATION_ID", "invalid station id")
}

func TestSimulationPermissionReturnsNotFoundWhenStationDoesNotExist(t *testing.T) {
	cfg := &config.Config{}
	cfg.Simulation.LineDefinitionPath = "internal/domain/simulation/fixtures/line.json"
	container := di.NewContainer(cfg)
	mux := setup(http.NewServeMux(), cfg, container)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/permission", strings.NewReader(`{"stationId":"S9","allowed":true}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	assertRouterErrorBody(t, rec.Body.Bytes(), "STATION_NOT_FOUND", "station not found")
}

func assertRouterErrorBody(t *testing.T, body []byte, code string, message string) {
	t.Helper()

	var got map[string]map[string]string
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if got["error"]["code"] != code {
		t.Fatalf("expected error code %s, got %q", code, got["error"]["code"])
	}
	if got["error"]["message"] != message {
		t.Fatalf("expected error message %s, got %q", message, got["error"]["message"])
	}
}
