package simulation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	simulationapp "github.com/right1121/railway-control-center-simulator/internal/application/simulation"
)

func TestGetReturnsSimulationDTO(t *testing.T) {
	uc := &stubSimulationUseCase{
		getDTO: testSimulationDTO(),
	}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulation", nil)
	rec := httptest.NewRecorder()
	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got simulationapp.SimulationDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if got.SimTimeMillis != 1000 {
		t.Fatalf("expected simTimeMillis 1000, got %d", got.SimTimeMillis)
	}
	if len(got.Trains) != 1 || got.Trains[0].ID != "T0" {
		t.Fatalf("unexpected trains payload: %+v", got.Trains)
	}
	if !got.Trains[0].AtBoundary {
		t.Fatalf("expected atBoundary true")
	}
	if got.Trains[0].StationID == nil || *got.Trains[0].StationID != "S1" {
		t.Fatalf("expected stationId S1, got %+v", got.Trains[0].StationID)
	}
}

func TestGetReturnsInternalErrorOnUseCaseFailure(t *testing.T) {
	uc := &stubSimulationUseCase{getErr: errors.New("boom")}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulation", nil)
	rec := httptest.NewRecorder()
	handler.Get(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	assertErrorBody(t, rec.Body.Bytes(), "INTERNAL", "internal error")
}

func TestTickReturnsSimulationDTO(t *testing.T) {
	uc := &stubSimulationUseCase{
		tickDTO: testSimulationDTO(),
	}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/tick", strings.NewReader(`{"deltaMillis":1000}`))
	rec := httptest.NewRecorder()
	handler.Tick(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if uc.tickInput.DeltaMillis != 1000 {
		t.Fatalf("expected deltaMillis 1000, got %d", uc.tickInput.DeltaMillis)
	}

	var got simulationapp.SimulationDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if got.SimTimeMillis != 1000 {
		t.Fatalf("expected simTimeMillis 1000, got %d", got.SimTimeMillis)
	}
	if !got.Trains[0].AtBoundary {
		t.Fatalf("expected atBoundary true")
	}
	if got.Trains[0].StationID == nil || *got.Trains[0].StationID != "S1" {
		t.Fatalf("expected stationId S1, got %+v", got.Trains[0].StationID)
	}
}

func TestTickReturnsBadJSONOnDecodeFailure(t *testing.T) {
	uc := &stubSimulationUseCase{}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/tick", strings.NewReader(`{"deltaMillis":`))
	rec := httptest.NewRecorder()
	handler.Tick(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorBody(t, rec.Body.Bytes(), "BAD_JSON", "invalid json")
}

func TestTickReturnsValidationErrorOnInvalidTickDelta(t *testing.T) {
	uc := &stubSimulationUseCase{tickErr: simulationapp.ErrInvalidTickDelta}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/tick", strings.NewReader(`{"deltaMillis":0}`))
	rec := httptest.NewRecorder()
	handler.Tick(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorBody(t, rec.Body.Bytes(), "INVALID_TICK_DELTA", "invalid tick delta")
}

func TestTickReturnsInternalErrorOnUseCaseFailure(t *testing.T) {
	uc := &stubSimulationUseCase{tickErr: errors.New("boom")}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/tick", strings.NewReader(`{"deltaMillis":1000}`))
	rec := httptest.NewRecorder()
	handler.Tick(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	assertErrorBody(t, rec.Body.Bytes(), "INTERNAL", "internal error")
}

func TestSetDeparturePermissionReturnsDTO(t *testing.T) {
	uc := &stubSimulationUseCase{
		permissionDTO: simulationapp.DeparturePermissionDTO{
			StationID: "S1",
			Allowed:   true,
		},
	}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/permission", strings.NewReader(`{"stationId":"S1","allowed":true}`))
	rec := httptest.NewRecorder()
	handler.SetDeparturePermission(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if uc.permissionInput.StationID != "S1" {
		t.Fatalf("expected stationId S1, got %s", uc.permissionInput.StationID)
	}
	if !uc.permissionInput.Allowed {
		t.Fatalf("expected allowed true")
	}

	var got simulationapp.DeparturePermissionDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if got.StationID != "S1" || !got.Allowed {
		t.Fatalf("unexpected permission response: %+v", got)
	}
}

func TestSetDeparturePermissionReturnsBadJSONOnDecodeFailure(t *testing.T) {
	uc := &stubSimulationUseCase{}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/permission", strings.NewReader(`{"stationId":`))
	rec := httptest.NewRecorder()
	handler.SetDeparturePermission(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorBody(t, rec.Body.Bytes(), "BAD_JSON", "invalid json")
}

func TestSetDeparturePermissionReturnsBadRequestOnInvalidStationID(t *testing.T) {
	uc := &stubSimulationUseCase{permissionErr: simulationapp.ErrInvalidStationID}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/permission", strings.NewReader(`{"stationId":" ","allowed":true}`))
	rec := httptest.NewRecorder()
	handler.SetDeparturePermission(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	assertErrorBody(t, rec.Body.Bytes(), "INVALID_STATION_ID", "invalid station id")
}

func TestSetDeparturePermissionReturnsNotFoundWhenStationDoesNotExist(t *testing.T) {
	uc := &stubSimulationUseCase{permissionErr: simulationapp.ErrStationNotFound}
	handler := NewSimulationHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/simulation/permission", strings.NewReader(`{"stationId":"S9","allowed":true}`))
	rec := httptest.NewRecorder()
	handler.SetDeparturePermission(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	assertErrorBody(t, rec.Body.Bytes(), "STATION_NOT_FOUND", "station not found")
}

type stubSimulationUseCase struct {
	getDTO          simulationapp.SimulationDTO
	tickDTO         simulationapp.SimulationDTO
	permissionDTO   simulationapp.DeparturePermissionDTO
	getErr          error
	tickErr         error
	permissionErr   error
	tickInput       simulationapp.TickInput
	permissionInput simulationapp.SetDeparturePermissionInput
}

func (s *stubSimulationUseCase) GetSimulation(ctx context.Context) (simulationapp.SimulationDTO, error) {
	_ = ctx
	return s.getDTO, s.getErr
}

func (s *stubSimulationUseCase) Tick(ctx context.Context, input simulationapp.TickInput) (simulationapp.SimulationDTO, error) {
	_ = ctx
	s.tickInput = input
	return s.tickDTO, s.tickErr
}

func (s *stubSimulationUseCase) SetDeparturePermission(ctx context.Context, input simulationapp.SetDeparturePermissionInput) (simulationapp.DeparturePermissionDTO, error) {
	_ = ctx
	s.permissionInput = input
	return s.permissionDTO, s.permissionErr
}

func testSimulationDTO() simulationapp.SimulationDTO {
	stationID := "S1"
	return simulationapp.SimulationDTO{
		SimTimeMillis: 1000,
		Line: simulationapp.LineDTO{
			Stations: []string{"S0", "S1"},
			Blocks:   []string{"B0"},
		},
		Trains: []simulationapp.TrainDTO{
			{
				ID:              "T0",
				BlockID:         "B0",
				Progress:        0.5,
				Forward:         true,
				Speed:           0.5,
				PendingTurnback: false,
				AtBoundary:      true,
				StationID:       &stationID,
			},
		},
	}
}

func assertErrorBody(t *testing.T, body []byte, code string, message string) {
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
