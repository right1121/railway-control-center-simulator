package filesystem

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	domain "github.com/right1121/railway-control-center-simulator/internal/domain/simulation"
)

func TestSimulationLineLoaderLoadValidJSON(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"},{"id":"S2"}],
  "blocks":[
    {"id":"B0","fromStationId":"S0","toStationId":"S1"},
    {"id":"B1","fromStationId":"S1","toStationId":"S2"}
  ]
}`)
	loader := NewSimulationLineLoader(path)

	line, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(line.Stations()) != 3 {
		t.Fatalf("expected 3 stations, got %d", len(line.Stations()))
	}
	if len(line.Blocks()) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(line.Blocks()))
	}
}

func TestSimulationLineLoaderLoadInvalidConnectivity(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"},{"id":"S2"}],
  "blocks":[
    {"id":"B0","fromStationId":"S0","toStationId":"S2"},
    {"id":"B1","fromStationId":"S1","toStationId":"S2"}
  ]
}`)
	loader := NewSimulationLineLoader(path)

	if _, err := loader.Load(context.Background()); err == nil {
		t.Fatalf("expected connectivity error")
	}
}

func TestSimulationLineLoaderUsesDefaultPathWhenEmpty(t *testing.T) {
	loader := NewSimulationLineLoader("   ")

	line, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(line.Stations()) != 3 {
		t.Fatalf("expected 3 stations from default fixture, got %d", len(line.Stations()))
	}
	if len(line.Blocks()) != 2 {
		t.Fatalf("expected 2 blocks from default fixture, got %d", len(line.Blocks()))
	}
}

func TestSimulationLineLoaderRejectsUnknownTopLevelField(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"}],
  "blocks":[{"id":"B0","fromStationId":"S0","toStationId":"S1"}],
  "extra":"x"
}`)
	loader := NewSimulationLineLoader(path)

	if _, err := loader.Load(context.Background()); err == nil {
		t.Fatalf("expected error for unknown top-level field")
	}
}

func TestSimulationLineLoaderRejectsUnknownNestedField(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0","name":"station-0"},{"id":"S1"}],
  "blocks":[{"id":"B0","fromStationId":"S0","toStationId":"S1"}]
}`)
	loader := NewSimulationLineLoader(path)

	if _, err := loader.Load(context.Background()); err == nil {
		t.Fatalf("expected error for unknown nested field")
	}
}

func TestSimulationLineLoaderRejectsMissingStations(t *testing.T) {
	path := writeFixture(t, `{
  "blocks":[{"id":"B0","fromStationId":"S0","toStationId":"S1"}]
}`)
	loader := NewSimulationLineLoader(path)

	if _, err := loader.Load(context.Background()); err == nil {
		t.Fatalf("expected missing stations error")
	}
}

func TestSimulationLineLoaderRejectsMissingBlocks(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"}]
}`)
	loader := NewSimulationLineLoader(path)

	if _, err := loader.Load(context.Background()); err == nil {
		t.Fatalf("expected missing blocks error")
	}
}

func TestSimulationLineLoaderRejectsEmptyStationID(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":" "}],
  "blocks":[{"id":"B0","fromStationId":"S0","toStationId":"S1"}]
}`)
	loader := NewSimulationLineLoader(path)

	assertLoadErrorIs(t, loader, domain.ErrStationIDEmpty)
}

func TestSimulationLineLoaderRejectsEmptyBlockID(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"}],
  "blocks":[{"id":" ","fromStationId":"S0","toStationId":"S1"}]
}`)
	loader := NewSimulationLineLoader(path)

	assertLoadErrorIs(t, loader, domain.ErrBlockIDEmpty)
}

func TestSimulationLineLoaderRejectsEmptyFromStationID(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"}],
  "blocks":[{"id":"B0","fromStationId":" ","toStationId":"S1"}]
}`)
	loader := NewSimulationLineLoader(path)

	assertLoadErrorIs(t, loader, domain.ErrStationIDEmpty)
}

func TestSimulationLineLoaderRejectsDuplicateStationID(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S0"},{"id":"S2"}],
  "blocks":[
    {"id":"B0","fromStationId":"S0","toStationId":"S0"},
    {"id":"B1","fromStationId":"S0","toStationId":"S2"}
  ]
}`)
	loader := NewSimulationLineLoader(path)

	assertLoadErrorIs(t, loader, domain.ErrLineDuplicateStationID)
}

func TestSimulationLineLoaderRejectsDuplicateBlockID(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"},{"id":"S2"}],
  "blocks":[
    {"id":"B0","fromStationId":"S0","toStationId":"S1"},
    {"id":"B0","fromStationId":"S1","toStationId":"S2"}
  ]
}`)
	loader := NewSimulationLineLoader(path)

	assertLoadErrorIs(t, loader, domain.ErrLineDuplicateBlockID)
}

func TestSimulationLineLoaderRejectsStationsBlocksMismatch(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"}],
  "blocks":[
    {"id":"B0","fromStationId":"S0","toStationId":"S1"},
    {"id":"B1","fromStationId":"S1","toStationId":"S2"}
  ]
}`)
	loader := NewSimulationLineLoader(path)

	assertLoadErrorIs(t, loader, domain.ErrLineStationsBlocksMismatch)
}

func TestSimulationLineLoaderRejectsNoBlocks(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"}],
  "blocks":[]
}`)
	loader := NewSimulationLineLoader(path)

	assertLoadErrorIs(t, loader, domain.ErrLineHasNoBlocks)
}

func TestSimulationLineLoaderRejectsConnectivityMismatch(t *testing.T) {
	path := writeFixture(t, `{
  "stations":[{"id":"S0"},{"id":"S1"},{"id":"S2"}],
  "blocks":[
    {"id":"B0","fromStationId":"S0","toStationId":"S1"},
    {"id":"B1","fromStationId":"S0","toStationId":"S2"}
  ]
}`)
	loader := NewSimulationLineLoader(path)

	assertLoadErrorIs(t, loader, domain.ErrLineConnectivityInvalid)
}

func assertLoadErrorIs(t *testing.T, loader *SimulationLineLoader, target error) {
	t.Helper()

	_, err := loader.Load(context.Background())
	if !errors.Is(err, target) {
		t.Fatalf("expected %v, got %v", target, err)
	}
}

func writeFixture(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "line.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture failed: %v", err)
	}
	return path
}
