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

func TestFromStationAndToStation(t *testing.T) {
	line := mustTestLine(t)
	b0 := mustBlockID(t, "B0")
	b1 := mustBlockID(t, "B1")
	b9 := mustBlockID(t, "B9")

	fromB0, ok := line.FromStation(b0)
	if !ok || fromB0.String() != "S0" {
		t.Fatalf("expected FromStation(B0)=S0, got ok=%t station=%s", ok, fromB0.String())
	}

	toB0, ok := line.ToStation(b0)
	if !ok || toB0.String() != "S1" {
		t.Fatalf("expected ToStation(B0)=S1, got ok=%t station=%s", ok, toB0.String())
	}

	fromB1, ok := line.FromStation(b1)
	if !ok || fromB1.String() != "S1" {
		t.Fatalf("expected FromStation(B1)=S1, got ok=%t station=%s", ok, fromB1.String())
	}

	toB1, ok := line.ToStation(b1)
	if !ok || toB1.String() != "S2" {
		t.Fatalf("expected ToStation(B1)=S2, got ok=%t station=%s", ok, toB1.String())
	}

	if _, ok := line.FromStation(b9); ok {
		t.Fatalf("expected FromStation(B9) to fail")
	}
	if _, ok := line.ToStation(b9); ok {
		t.Fatalf("expected ToStation(B9) to fail")
	}
}

func mustTestLine(t *testing.T) *Line {
	t.Helper()

	line, err := NewLine(
		[]StationID{
			mustStationID(t, "S0"),
			mustStationID(t, "S1"),
			mustStationID(t, "S2"),
		},
		[]BlockID{
			mustBlockID(t, "B0"),
			mustBlockID(t, "B1"),
		},
	)
	if err != nil {
		t.Fatalf("new line failed: %v", err)
	}
	return line
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
