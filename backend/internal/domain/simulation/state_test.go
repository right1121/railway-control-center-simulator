package simulation

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTickMovesInsideBlock(t *testing.T) {
	state := newStateWithBlocks(t, 2)
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
	state := newStateWithBlocks(t, 2)
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

func TestTickLargeDeltaCrossesMultipleBlocks(t *testing.T) {
	state := newStateWithBlocks(t, 4)
	train := newTestTrain(t, "T0", "B0", 0.0, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	delta, _ := NewTickDelta(5 * time.Second)
	if err := state.Tick(delta); err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	got := state.Trains()[0]
	if got.BlockID().String() != "B2" {
		t.Fatalf("expected block B2, got %s", got.BlockID().String())
	}
	if got.Progress().Float64() != 0.5 {
		t.Fatalf("expected progress 0.5, got %f", got.Progress().Float64())
	}
}

func TestTickTerminalTurnbackNextTick(t *testing.T) {
	state := newStateWithBlocks(t, 2)
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
	state := newStateWithBlocks(t, 2)
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

func TestTickCanEnterFreedBlockOnNextTick(t *testing.T) {
	state := newStateWithBlocks(t, 3)
	lead := newTestTrain(t, "T0", "B0", 0.9, true, 0.5)
	blocker := newTestTrain(t, "T1", "B1", 0.9, true, 0.5)

	if err := state.AddTrain(lead); err != nil {
		t.Fatalf("add lead failed: %v", err)
	}
	if err := state.AddTrain(blocker); err != nil {
		t.Fatalf("add blocker failed: %v", err)
	}

	delta, _ := NewTickDelta(time.Second)
	if err := state.Tick(delta); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}

	first := state.Trains()[0]
	if first.BlockID().String() != "B0" || first.Progress().Float64() != 1.0 {
		t.Fatalf("expected T0 waiting at B0 boundary after first tick, got block=%s progress=%f", first.BlockID().String(), first.Progress().Float64())
	}

	if err := state.Tick(delta); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}

	second := state.Trains()[0]
	if second.BlockID().String() != "B1" {
		t.Fatalf("expected T0 to enter B1 on second tick, got %s", second.BlockID().String())
	}
}

func TestTickDoesNotDepartAtBoundaryWhenDeparturePermissionIsOff(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	train := newTestTrain(t, "T0", "B0", 1.0, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	if err := setDeparturePermissionForTest(t, state, "S1", false); err != nil {
		t.Fatalf("set departure permission failed: %v", err)
	}

	delta, _ := NewTickDelta(time.Second)
	if err := state.Tick(delta); err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	got := state.Trains()[0]
	if got.BlockID().String() != "B0" {
		t.Fatalf("expected block B0, got %s", got.BlockID().String())
	}
	if got.Progress().Float64() != 1.0 {
		t.Fatalf("expected progress 1.0, got %f", got.Progress().Float64())
	}
}

func TestTickDepartsOnNextTickAfterDeparturePermissionTurnsOn(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	train := newTestTrain(t, "T0", "B0", 1.0, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	if err := setDeparturePermissionForTest(t, state, "S1", false); err != nil {
		t.Fatalf("set departure permission failed: %v", err)
	}

	delta, _ := NewTickDelta(time.Second)
	if err := state.Tick(delta); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}

	waiting := state.Trains()[0]
	if waiting.BlockID().String() != "B0" || waiting.Progress().Float64() != 1.0 {
		t.Fatalf("expected waiting at B0 boundary, got block=%s progress=%f", waiting.BlockID().String(), waiting.Progress().Float64())
	}

	if err := setDeparturePermissionForTest(t, state, "S1", true); err != nil {
		t.Fatalf("enable departure permission failed: %v", err)
	}

	if err := state.Tick(delta); err != nil {
		t.Fatalf("second tick failed: %v", err)
	}

	got := state.Trains()[0]
	if got.BlockID().String() != "B1" {
		t.Fatalf("expected block B1, got %s", got.BlockID().String())
	}
	if got.Progress().Float64() != 0.5 {
		t.Fatalf("expected progress 0.5, got %f", got.Progress().Float64())
	}
}

func TestTickDoesNotEnterOccupiedNextBlockEvenWhenDeparturePermissionIsOn(t *testing.T) {
	state := newStateWithBlocks(t, 3)
	lead := newTestTrain(t, "T0", "B0", 1.0, true, 0.5)
	blocker := newTestTrain(t, "T1", "B1", 0.4, true, 0.5)

	if err := state.AddTrain(lead); err != nil {
		t.Fatalf("add lead failed: %v", err)
	}
	if err := state.AddTrain(blocker); err != nil {
		t.Fatalf("add blocker failed: %v", err)
	}

	if err := setDeparturePermissionForTest(t, state, "S1", true); err != nil {
		t.Fatalf("set departure permission failed: %v", err)
	}

	delta, _ := NewTickDelta(time.Second)
	if err := state.Tick(delta); err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	got := state.Trains()[0]
	if got.ID().String() != "T0" {
		t.Fatalf("expected first sorted train T0, got %s", got.ID().String())
	}
	if got.BlockID().String() != "B0" {
		t.Fatalf("expected T0 to stay in B0, got %s", got.BlockID().String())
	}
	if got.Progress().Float64() != 1.0 {
		t.Fatalf("expected T0 clamped at boundary, got %f", got.Progress().Float64())
	}
}

func TestSetDeparturePermissionRejectsUnknownStation(t *testing.T) {
	state := newStateWithBlocks(t, 2)

	err := setDeparturePermissionForTest(t, state, "S9", true)
	if err == nil {
		t.Fatalf("expected error for unknown station")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestTickRejectsNonPositiveDelta(t *testing.T) {
	state := newStateWithBlocks(t, 1)
	train := newTestTrain(t, "T0", "B0", 0.0, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	if err := state.Tick(TickDelta{}); err != ErrTickDeltaNotPositive {
		t.Fatalf("expected ErrTickDeltaNotPositive, got %v", err)
	}
}

func TestTickDeterministicAcrossInsertionOrder(t *testing.T) {
	makeState := func(order []string) *SimulationState {
		t.Helper()
		state := newStateWithBlocks(t, 4)
		for _, id := range order {
			switch id {
			case "T0":
				if err := state.AddTrain(newTestTrain(t, "T0", "B0", 0.25, true, 0.5)); err != nil {
					t.Fatalf("add T0 failed: %v", err)
				}
			case "T1":
				if err := state.AddTrain(newTestTrain(t, "T1", "B2", 0.75, false, 0.5)); err != nil {
					t.Fatalf("add T1 failed: %v", err)
				}
			default:
				t.Fatalf("unknown train id: %s", id)
			}
		}
		return state
	}

	left := makeState([]string{"T0", "T1"})
	right := makeState([]string{"T1", "T0"})

	delta, _ := NewTickDelta(1500 * time.Millisecond)
	if err := left.Tick(delta); err != nil {
		t.Fatalf("left tick failed: %v", err)
	}
	if err := right.Tick(delta); err != nil {
		t.Fatalf("right tick failed: %v", err)
	}

	if snapshot(left) != snapshot(right) {
		t.Fatalf("expected deterministic state, left=%q right=%q", snapshot(left), snapshot(right))
	}
}

func TestTrainStationAtBoundaryReturnsToStationForForwardAtProgressOne(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	train := newTestTrain(t, "T0", "B0", 1.0, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	station, ok := state.TrainStationAtBoundary(train.ID())
	if !ok {
		t.Fatalf("expected boundary station")
	}
	if station.String() != "S1" {
		t.Fatalf("expected S1, got %s", station.String())
	}
}

func TestTrainStationAtBoundaryReturnsFromStationForBackwardAtProgressZero(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	train := newTestTrain(t, "T0", "B1", 0.0, false, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	station, ok := state.TrainStationAtBoundary(train.ID())
	if !ok {
		t.Fatalf("expected boundary station")
	}
	if station.String() != "S1" {
		t.Fatalf("expected S1, got %s", station.String())
	}
}

func TestTrainStationAtBoundaryReturnsFalseWhenDirectionBoundaryDoesNotMatch(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	train := newTestTrain(t, "T0", "B0", 0.0, true, 0.5)
	if err := state.AddTrain(train); err != nil {
		t.Fatalf("add train failed: %v", err)
	}

	if !state.IsAtBoundary(train.ID()) {
		t.Fatalf("expected IsAtBoundary true")
	}
	if _, ok := state.TrainStationAtBoundary(train.ID()); ok {
		t.Fatalf("expected no station for forward train at progress 0.0")
	}
}

func TestTrainStationAtBoundaryReturnsFalseForUnknownTrain(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	unknown, _ := NewTrainID("UNKNOWN")

	if state.IsAtBoundary(unknown) {
		t.Fatalf("expected IsAtBoundary false for unknown train")
	}
	if _, ok := state.TrainStation(unknown); ok {
		t.Fatalf("expected TrainStation false for unknown train")
	}
	if _, ok := state.TrainStationAtBoundary(unknown); ok {
		t.Fatalf("expected TrainStationAtBoundary false for unknown train")
	}
}

func TestAddTrainRejectsDuplicateID(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	first := newTestTrain(t, "T0", "B0", 0.1, true, 0.5)
	second := newTestTrain(t, "T0", "B1", 0.1, true, 0.5)

	if err := state.AddTrain(first); err != nil {
		t.Fatalf("add first train failed: %v", err)
	}
	if err := state.AddTrain(second); err != ErrTrainAlreadyExists {
		t.Fatalf("expected ErrTrainAlreadyExists, got %v", err)
	}
}

func TestAddTrainRejectsUnknownBlock(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	train := newTestTrain(t, "T0", "B9", 0.1, true, 0.5)

	if err := state.AddTrain(train); err != ErrBlockNotFound {
		t.Fatalf("expected ErrBlockNotFound, got %v", err)
	}
}

func TestAddTrainRejectsOccupiedBlock(t *testing.T) {
	state := newStateWithBlocks(t, 2)
	first := newTestTrain(t, "T0", "B0", 0.1, true, 0.5)
	second := newTestTrain(t, "T1", "B0", 0.2, true, 0.5)

	if err := state.AddTrain(first); err != nil {
		t.Fatalf("add first train failed: %v", err)
	}
	if err := state.AddTrain(second); err != ErrBlockOccupied {
		t.Fatalf("expected ErrBlockOccupied, got %v", err)
	}
}

func TestAddTrainRejectsNilTrain(t *testing.T) {
	state := newStateWithBlocks(t, 2)

	if err := state.AddTrain(nil); err != ErrTrainNotFound {
		t.Fatalf("expected ErrTrainNotFound, got %v", err)
	}
}

func newStateWithBlocks(t *testing.T, blockCount int) *SimulationState {
	t.Helper()

	stations := make([]StationID, 0, blockCount+1)
	for i := 0; i <= blockCount; i++ {
		id, err := NewStationID(fmt.Sprintf("S%d", i))
		if err != nil {
			t.Fatalf("new station failed: %v", err)
		}
		stations = append(stations, id)
	}

	blocks := make([]BlockID, 0, blockCount)
	for i := 0; i < blockCount; i++ {
		id, err := NewBlockID(fmt.Sprintf("B%d", i))
		if err != nil {
			t.Fatalf("new block failed: %v", err)
		}
		blocks = append(blocks, id)
	}

	line, err := NewLine(stations, blocks)
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

func snapshot(state *SimulationState) string {
	trains := state.Trains()
	out := fmt.Sprintf("sim=%d;", state.SimTime().Millis())
	for _, train := range trains {
		out += fmt.Sprintf("%s:%s:%.9f:%t:%t;", train.ID().String(), train.BlockID().String(), train.Progress().Float64(), train.Forward(), train.PendingTurnback())
	}
	return out
}

func setDeparturePermissionForTest(t *testing.T, state *SimulationState, stationID string, allowed bool) error {
	t.Helper()

	method := reflect.ValueOf(state).MethodByName("SetDeparturePermission")
	if !method.IsValid() {
		t.Fatalf("SimulationState must implement SetDeparturePermission(StationID, bool) error")
	}

	methodType := method.Type()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if methodType.NumIn() != 2 || methodType.In(0) != reflect.TypeOf(StationID{}) || methodType.In(1).Kind() != reflect.Bool {
		t.Fatalf("SetDeparturePermission signature must be (StationID, bool), got %s", methodType.String())
	}
	if methodType.NumOut() != 1 || !methodType.Out(0).Implements(errorType) {
		t.Fatalf("SetDeparturePermission must return error, got %s", methodType.String())
	}

	station, err := NewStationID(stationID)
	if err != nil {
		t.Fatalf("new station id failed: %v", err)
	}

	results := method.Call([]reflect.Value{reflect.ValueOf(station), reflect.ValueOf(allowed)})
	if results[0].IsNil() {
		return nil
	}

	callErr, ok := results[0].Interface().(error)
	if !ok {
		t.Fatalf("SetDeparturePermission returned non-error value: %T", results[0].Interface())
	}
	return callErr
}
