package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromPathParsesSimulationLineDefinitionPath(t *testing.T) {
	path := writeConfigFixture(t, `{
  "environment":"development",
  "server":{"port":8080,"host":"localhost"},
  "simulation":{"lineDefinitionPath":"fixtures/custom-line.json"}
}`)

	cfg, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if cfg.Simulation.LineDefinitionPath != "fixtures/custom-line.json" {
		t.Fatalf("expected lineDefinitionPath fixtures/custom-line.json, got %q", cfg.Simulation.LineDefinitionPath)
	}
}

func TestLoadFromPathKeepsSimulationLineDefinitionPathEmptyWhenMissing(t *testing.T) {
	path := writeConfigFixture(t, `{
  "environment":"development",
  "server":{"port":8080,"host":"localhost"}
}`)

	cfg, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if cfg.Simulation.LineDefinitionPath != "" {
		t.Fatalf("expected empty lineDefinitionPath, got %q", cfg.Simulation.LineDefinitionPath)
	}
}

func writeConfigFixture(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture failed: %v", err)
	}
	return path
}
