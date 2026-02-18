package simulation

import (
	"encoding/json"
	"errors"
	"net/http"

	simulationapp "github.com/right1121/railway-control-center-simulator/internal/application/simulation"
	"github.com/right1121/railway-control-center-simulator/internal/interfaces/http/utils"
)

type SimulationHandler struct {
	usecase simulationapp.UseCase
}

func NewSimulationHandler(uc simulationapp.UseCase) *SimulationHandler {
	return &SimulationHandler{usecase: uc}
}

func (h *SimulationHandler) Get(w http.ResponseWriter, r *http.Request) {
	dto, err := h.usecase.GetSimulation(r.Context())
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.ErrBody("INTERNAL", "internal error"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, dto)
}

type tickReq struct {
	DeltaMillis int64 `json:"deltaMillis"`
}

type permissionReq struct {
	StationID string `json:"stationId"`
	Allowed   bool   `json:"allowed"`
}

func (h *SimulationHandler) Tick(w http.ResponseWriter, r *http.Request) {
	var req tickReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.BadJSON())
		return
	}

	dto, err := h.usecase.Tick(r.Context(), simulationapp.TickInput{
		DeltaMillis: req.DeltaMillis,
	})
	if err != nil {
		if errors.Is(err, simulationapp.ErrInvalidTickDelta) {
			utils.WriteJSON(w, http.StatusBadRequest, utils.ErrBody("INVALID_TICK_DELTA", "invalid tick delta"))
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.ErrBody("INTERNAL", "internal error"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, dto)
}

func (h *SimulationHandler) SetDeparturePermission(w http.ResponseWriter, r *http.Request) {
	var req permissionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.BadJSON())
		return
	}

	dto, err := h.usecase.SetDeparturePermission(r.Context(), simulationapp.SetDeparturePermissionInput{
		StationID: req.StationID,
		Allowed:   req.Allowed,
	})
	if err != nil {
		if errors.Is(err, simulationapp.ErrInvalidStationID) {
			utils.WriteJSON(w, http.StatusBadRequest, utils.ErrBody("INVALID_STATION_ID", "invalid station id"))
			return
		}
		if errors.Is(err, simulationapp.ErrStationNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, utils.ErrBody("STATION_NOT_FOUND", "station not found"))
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, utils.ErrBody("INTERNAL", "internal error"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, dto)
}
