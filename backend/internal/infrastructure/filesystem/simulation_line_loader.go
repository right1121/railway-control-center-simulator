package filesystem

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	data, err := readLineData(l.path)
	if err != nil {
		return nil, err
	}

	raw, err := decodeLineJSON(data)
	if err != nil {
		return nil, err
	}

	if raw.Stations == nil {
		return nil, fmt.Errorf("line fixture parse failed: stations is required")
	}
	if raw.Blocks == nil {
		return nil, fmt.Errorf("line fixture parse failed: blocks is required")
	}
	if len(raw.Stations) != len(raw.Blocks)+1 {
		return nil, domain.ErrLineStationsBlocksMismatch
	}

	stations := make([]domain.StationID, len(raw.Stations))
	stationSeen := make(map[string]struct{}, len(raw.Stations))
	for i, s := range raw.Stations {
		id, err := domain.NewStationID(s.ID)
		if err != nil {
			return nil, err
		}
		key := id.String()
		if _, exists := stationSeen[key]; exists {
			return nil, domain.ErrLineDuplicateStationID
		}
		stationSeen[key] = struct{}{}
		stations[i] = id
	}

	blocks := make([]domain.BlockID, len(raw.Blocks))
	blockSeen := make(map[string]struct{}, len(raw.Blocks))
	for i, b := range raw.Blocks {
		id, err := domain.NewBlockID(b.ID)
		if err != nil {
			return nil, err
		}
		key := id.String()
		if _, exists := blockSeen[key]; exists {
			return nil, domain.ErrLineDuplicateBlockID
		}
		blockSeen[key] = struct{}{}

		fromStationID, err := domain.NewStationID(b.FromStationID)
		if err != nil {
			return nil, err
		}
		toStationID, err := domain.NewStationID(b.ToStationID)
		if err != nil {
			return nil, err
		}
		if fromStationID.String() != stations[i].String() {
			return nil, domain.ErrLineConnectivityInvalid
		}
		if toStationID.String() != stations[i+1].String() {
			return nil, domain.ErrLineConnectivityInvalid
		}

		blocks[i] = id
	}

	line, err := domain.NewLine(stations, blocks)
	if err != nil {
		return nil, err
	}

	return line, nil
}

func readLineData(path string) ([]byte, error) {
	candidates := buildCandidatePaths(path)
	var lastErr error
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, fmt.Errorf("line fixture read failed: %w", lastErr)
	}
	return nil, fmt.Errorf("line fixture read failed: no candidate path")
}

func buildCandidatePaths(path string) []string {
	normalized := filepath.ToSlash(path)
	alternatives := []string{normalized}
	if strings.HasPrefix(normalized, "backend/") {
		alternatives = append(alternatives, strings.TrimPrefix(normalized, "backend/"))
	}

	candidates := make([]string, 0, 16)
	seen := make(map[string]struct{}, 16)
	add := func(candidate string) {
		if candidate == "" {
			return
		}
		clean := filepath.Clean(candidate)
		if _, exists := seen[clean]; exists {
			return
		}
		seen[clean] = struct{}{}
		candidates = append(candidates, clean)
	}

	for _, candidate := range alternatives {
		add(candidate)
	}

	wd, err := os.Getwd()
	if err != nil {
		return candidates
	}
	base := wd
	for i := 0; i < 8; i++ {
		for _, candidate := range alternatives {
			add(filepath.Join(base, candidate))
		}

		parent := filepath.Dir(base)
		if parent == base {
			break
		}
		base = parent
	}

	return candidates
}

func decodeLineJSON(data []byte) (simulationLineJSON, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var raw simulationLineJSON
	if err := decoder.Decode(&raw); err != nil {
		return simulationLineJSON{}, fmt.Errorf("line fixture parse failed: %w", err)
	}

	// Reject trailing non-whitespace tokens after the first JSON object.
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return simulationLineJSON{}, fmt.Errorf("line fixture parse failed: unexpected trailing data")
	}

	return raw, nil
}
