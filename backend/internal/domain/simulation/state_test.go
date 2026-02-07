package simulation

import (
	"testing"
	"time"
)

func TestTickMovesInsideBlock(t *testing.T) {
	state := newTestState(t)
	train := newTestTrain(t, "T0", "B0", 0.0, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	delta, _ := NewTickDelta(time.Second)
	if err := state.Tick(delta); err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	got := state.Trains()[0]
	if got.BlockID().String() != "B0" {
		t.Fatalf("expected block B0, got %s", got.BlockID().String())
	}
	if got.Progress().Float64() != 0.5 {
		t.Fatalf("expected progress 0.5, got %f", got.Progress().Float64())
	}
}

func TestTickCrossesBoundary(t *testing.T) {
	state := newTestState(t)
	train := newTestTrain(t, "T0", "B0", 0.0, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	delta, _ := NewTickDelta(3 * time.Second)
	if err := state.Tick(delta); err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	got := state.Trains()[0]
	if got.BlockID().String() != "B1" {
		t.Fatalf("expected block B1, got %s", got.BlockID().String())
	}
	if got.Progress().Float64() != 0.5 {
		t.Fatalf("expected progress 0.5, got %f", got.Progress().Float64())
	}
}

func TestTickTerminalTurnbackNextTick(t *testing.T) {
	state := newTestState(t)
	train := newTestTrain(t, "T0", "B1", 0.9, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	first, _ := NewTickDelta(time.Second)
	if err := state.Tick(first); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}

	got := state.Trains()[0]
	if got.BlockID().String() != "B1" {
		t.Fatalf("expected block B1, got %s", got.BlockID().String())
	}
	if got.Progress().Float64() != 1.0 {
		t.Fatalf("expected progress 1.0, got %f", got.Progress().Float64())
	}
	if !got.PendingTurnback() {
		t.Fatalf("expected pending turnback true")
	}

	second, _ := NewTickDelta(time.Second)
	if err := state.Tick(second); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}

	got = state.Trains()[0]
	if got.Forward() {
		t.Fatalf("expected direction to be backward after turnback")
	}
	if got.Progress().Float64() != 0.5 {
		t.Fatalf("expected progress 0.5 after reverse move, got %f", got.Progress().Float64())
	}
	if got.PendingTurnback() {
		t.Fatalf("expected pending turnback false")
	}
}

func TestTickBlocksOccupiedNextBlock(t *testing.T) {
	state := newTestState(t)
	lead := newTestTrain(t, "T0", "B0", 0.9, true, 0.5)
	blocker := newTestTrain(t, "T1", "B1", 0.5, true, 0.5)

	if err := state.AddTrain(lead); err != nil {
		t.Fatalf("add lead failed: %v", err)
	}
	if err := state.AddTrain(blocker); err != nil {
		t.Fatalf("add blocker failed: %v", err)
	}

	delta, _ := NewTickDelta(time.Second)
	if err := state.Tick(delta); err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	trains := state.Trains()
	if trains[0].ID().String() != "T0" {
		t.Fatalf("expected sorted train order")
	}
	if trains[0].BlockID().String() != "B0" {
		t.Fatalf("expected T0 stay in B0, got %s", trains[0].BlockID().String())
	}
	if trains[0].Progress().Float64() != 1.0 {
		t.Fatalf("expected T0 clamped at boundary, got %f", trains[0].Progress().Float64())
	}
}

func newTestState(t *testing.T) *SimulationState {
	t.Helper()

	s0, _ := NewStationID("S0")
	s1, _ := NewStationID("S1")
	s2, _ := NewStationID("S2")
	b0, _ := NewBlockID("B0")
	b1, _ := NewBlockID("B1")

	line, err := NewLine([]StationID{s0, s1, s2}, []BlockID{b0, b1})
	if err != nil {
		t.Fatalf("new line failed: %v", err)
	}

	state, err := NewSimulationState(line)
	if err != nil {
		t.Fatalf("new state failed: %v", err)
	}
	return state
}

func newTestTrain(t *testing.T, trainID string, blockID string, progress float64, forward bool, speed float64) *Train {
	t.Helper()

	id, _ := NewTrainID(trainID)
	block, _ := NewBlockID(blockID)
	p, _ := NewBlockProgress(progress)
	train, err := NewTrain(id, block, p, forward, speed)
	if err != nil {
		t.Fatalf("new train failed: %v", err)
	}
	return train
}
