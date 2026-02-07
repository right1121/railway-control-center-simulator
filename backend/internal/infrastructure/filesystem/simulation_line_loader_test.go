package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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

func writeFixture(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "line.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture failed: %v", err)
	}
	return path
}
