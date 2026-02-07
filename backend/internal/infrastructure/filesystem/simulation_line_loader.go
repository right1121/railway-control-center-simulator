package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	domain "github.com/right1121/railway-control-center-simulator/internal/domain/simulation"
)

const DefaultSimulationLinePath = "backend/internal/domain/simulation/fixtures/line.json"

type SimulationLineLoader struct {
	path string
}

func NewSimulationLineLoader(path string) *SimulationLineLoader {
	path = strings.TrimSpace(path)
	if path == "" {
		path = DefaultSimulationLinePath
	}
	return &SimulationLineLoader{path: path}
}

type simulationLineJSON struct {
	Stations []stationJSON `json:"stations"`
	Blocks   []blockJSON   `json:"blocks"`
}

type stationJSON struct {
	ID string `json:"id"`
}

type blockJSON struct {
	ID            string `json:"id"`
	FromStationID string `json:"fromStationId"`
	ToStationID   string `json:"toStationId"`
}

func (l *SimulationLineLoader) Load(ctx context.Context) (*domain.Line, error) {
	_ = ctx

	data, err := os.ReadFile(l.path)
	if err != nil && strings.HasPrefix(l.path, "backend/") {
		altPath := strings.TrimPrefix(l.path, "backend/")
		data, err = os.ReadFile(altPath)
	}
	if err != nil {
		return nil, fmt.Errorf("line fixture read failed: %w", err)
	}

	var raw simulationLineJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("line fixture parse failed: %w", err)
	}

	stations := make([]domain.StationID, 0, len(raw.Stations))
	for _, s := range raw.Stations {
		id, err := domain.NewStationID(s.ID)
		if err != nil {
			return nil, err
		}
		stations = append(stations, id)
	}

	blocks := make([]domain.BlockID, 0, len(raw.Blocks))
	for _, b := range raw.Blocks {
		id, err := domain.NewBlockID(b.ID)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, id)
	}

	line, err := domain.NewLine(stations, blocks)
	if err != nil {
		return nil, err
	}

	if len(raw.Blocks) != len(raw.Stations)-1 {
		return nil, domain.ErrLineConnectivityInvalid
	}
	for i, block := range raw.Blocks {
		if strings.TrimSpace(block.FromStationID) != raw.Stations[i].ID {
			return nil, domain.ErrLineConnectivityInvalid
		}
		if strings.TrimSpace(block.ToStationID) != raw.Stations[i+1].ID {
			return nil, domain.ErrLineConnectivityInvalid
		}
	}

	return line, nil
}
