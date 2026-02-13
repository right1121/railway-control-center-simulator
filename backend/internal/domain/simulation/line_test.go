package simulation

import (
	"errors"
	"testing"
)

func TestNewLineRejectsNoBlocks(t *testing.T) {
	stations := []StationID{
		mustStationID(t, "S0"),
	}

	_, err := NewLine(stations, nil)
	if !errors.Is(err, ErrLineHasNoBlocks) {
		t.Fatalf("expected ErrLineHasNoBlocks, got %v", err)
	}
}

func TestNewLineRejectsStationsBlocksMismatch(t *testing.T) {
	stations := []StationID{
		mustStationID(t, "S0"),
		mustStationID(t, "S1"),
	}
	blocks := []BlockID{
		mustBlockID(t, "B0"),
		mustBlockID(t, "B1"),
	}

	_, err := NewLine(stations, blocks)
	if !errors.Is(err, ErrLineStationsBlocksMismatch) {
		t.Fatalf("expected ErrLineStationsBlocksMismatch, got %v", err)
	}
}

func TestNewLineRejectsDuplicateStationID(t *testing.T) {
	stations := []StationID{
		mustStationID(t, "S0"),
		mustStationID(t, "S0"),
		mustStationID(t, "S2"),
	}
	blocks := []BlockID{
		mustBlockID(t, "B0"),
		mustBlockID(t, "B1"),
	}

	_, err := NewLine(stations, blocks)
	if !errors.Is(err, ErrLineDuplicateStationID) {
		t.Fatalf("expected ErrLineDuplicateStationID, got %v", err)
	}
}

func TestNewLineRejectsDuplicateBlockID(t *testing.T) {
	stations := []StationID{
		mustStationID(t, "S0"),
		mustStationID(t, "S1"),
		mustStationID(t, "S2"),
	}
	blocks := []BlockID{
		mustBlockID(t, "B0"),
		mustBlockID(t, "B0"),
	}

	_, err := NewLine(stations, blocks)
	if !errors.Is(err, ErrLineDuplicateBlockID) {
		t.Fatalf("expected ErrLineDuplicateBlockID, got %v", err)
	}
}

func mustStationID(t *testing.T, raw string) StationID {
	t.Helper()
	id, err := NewStationID(raw)
	if err != nil {
		t.Fatalf("station id build failed: %v", err)
	}
	return id
}

func mustBlockID(t *testing.T, raw string) BlockID {
	t.Helper()
	id, err := NewBlockID(raw)
	if err != nil {
		t.Fatalf("block id build failed: %v", err)
	}
	return id
}
