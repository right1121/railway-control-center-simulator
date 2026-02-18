package simulation

import (
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	domain "github.com/right1121/railway-control-center-simulator/internal/domain/simulation"
	"github.com/right1121/railway-control-center-simulator/internal/infrastructure/memory"
)

func TestGetSimulationCreatesStateOnFirstCall(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	dto, err := uc.GetSimulation(context.Background())
	if err != nil {
		t.Fatalf("GetSimulation failed: %v", err)
	}

	if dto.SimTimeMillis != 0 {
		t.Fatalf("expected sim time 0, got %d", dto.SimTimeMillis)
	}
	if len(dto.Trains) != 1 {
		t.Fatalf("expected initial 1 train, got %d", len(dto.Trains))
	}
	if dto.Trains[0].BlockID != "B0" {
		t.Fatalf("expected initial block B0, got %s", dto.Trains[0].BlockID)
	}
	if !dto.Trains[0].AtBoundary {
		t.Fatalf("expected initial train at boundary")
	}
	if dto.Trains[0].StationID != nil {
		t.Fatalf("expected initial stationId nil for forward train at progress 0")
	}
	if len(dto.Line.Stations) != 3 {
		t.Fatalf("expected 3 stations, got %d", len(dto.Line.Stations))
	}
	if len(dto.Line.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(dto.Line.Blocks))
	}
}

func TestGetSimulationReturnsErrorOnLineLoadFailure(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	uc := NewUseCase(repo, &stubLineLoader{err: errors.New("broken json")})

	if _, err := uc.GetSimulation(context.Background()); err == nil {
		t.Fatalf("expected error on line load failure")
	}
}

func TestTickAdvancesSimulation(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	dto, err := uc.Tick(context.Background(), TickInput{DeltaMillis: 1000})
	if err != nil {
		t.Fatalf("Tick failed: %v", err)
	}

	if dto.SimTimeMillis != 1000 {
		t.Fatalf("expected sim time 1000ms, got %d", dto.SimTimeMillis)
	}
	if dto.Trains[0].Progress != 0.5 {
		t.Fatalf("expected train progress 0.5, got %f", dto.Trains[0].Progress)
	}
	if dto.Trains[0].AtBoundary {
		t.Fatalf("expected train not at boundary")
	}
	if dto.Trains[0].StationID != nil {
		t.Fatalf("expected stationId nil when not at boundary")
	}
}

func TestTickSetsBoundaryStationInDTO(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	dto, err := uc.Tick(context.Background(), TickInput{DeltaMillis: 4000})
	if err != nil {
		t.Fatalf("Tick failed: %v", err)
	}

	train := dto.Trains[0]
	if !train.AtBoundary {
		t.Fatalf("expected train at boundary")
	}
	if train.StationID == nil {
		t.Fatalf("expected stationId not nil")
	}
	if *train.StationID != "S2" {
		t.Fatalf("expected stationId S2, got %s", *train.StationID)
	}
}

func TestTickReturnsValidationErrorWhenDeltaIsNotPositive(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, err := uc.Tick(context.Background(), TickInput{DeltaMillis: 0})
	if !errors.Is(err, ErrInvalidTickDelta) {
		t.Fatalf("expected ErrInvalidTickDelta, got %v", err)
	}
}

func TestTickReturnsValidationErrorWhenDeltaOverflowsDuration(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, err := uc.Tick(context.Background(), TickInput{DeltaMillis: math.MaxInt64})
	if !errors.Is(err, ErrInvalidTickDelta) {
		t.Fatalf("expected ErrInvalidTickDelta, got %v", err)
	}
}

func TestSetDeparturePermissionControlsDepartureAtBoundary(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	stationID, allowed, err := callSetDeparturePermissionForTest(t, uc, "S1", false)
	if err != nil {
		t.Fatalf("SetDeparturePermission failed: %v", err)
	}
	if stationID != "S1" {
		t.Fatalf("expected stationId S1, got %s", stationID)
	}
	if allowed {
		t.Fatalf("expected allowed false")
	}

	if _, err := uc.Tick(context.Background(), TickInput{DeltaMillis: 2000}); err != nil {
		t.Fatalf("first tick failed: %v", err)
	}

	waiting, err := uc.Tick(context.Background(), TickInput{DeltaMillis: 1000})
	if err != nil {
		t.Fatalf("second tick failed: %v", err)
	}
	if waiting.Trains[0].BlockID != "B0" || waiting.Trains[0].Progress != 1.0 {
		t.Fatalf("expected waiting at B0 boundary, got block=%s progress=%f", waiting.Trains[0].BlockID, waiting.Trains[0].Progress)
	}

	if _, _, err := callSetDeparturePermissionForTest(t, uc, "S1", true); err != nil {
		t.Fatalf("enable departure permission failed: %v", err)
	}

	moved, err := uc.Tick(context.Background(), TickInput{DeltaMillis: 1000})
	if err != nil {
		t.Fatalf("third tick failed: %v", err)
	}
	if moved.Trains[0].BlockID != "B1" || moved.Trains[0].Progress != 0.5 {
		t.Fatalf("expected moved to B1 with progress 0.5, got block=%s progress=%f", moved.Trains[0].BlockID, moved.Trains[0].Progress)
	}
}

func TestSetDeparturePermissionReturnsValidationErrorWhenStationIDIsInvalid(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, _, err := callSetDeparturePermissionForTest(t, uc, " ", true)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestSetDeparturePermissionReturnsNotFoundWhenStationDoesNotExist(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, _, err := callSetDeparturePermissionForTest(t, uc, "S9", true)
	if err == nil {
		t.Fatalf("expected not found error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestAddTrainReturnsSimulationDTOWithStableOrder(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	dto, err := callAddTrainForTest(t, uc, "T1", "B1", 0.0, "Down", 0.5)
	if err != nil {
		t.Fatalf("AddTrain failed: %v", err)
	}

	if len(dto.Trains) != 2 {
		t.Fatalf("expected 2 trains, got %d", len(dto.Trains))
	}
	if dto.Trains[0].ID != "T0" || dto.Trains[1].ID != "T1" {
		t.Fatalf("expected stable sorted order [T0,T1], got [%s,%s]", dto.Trains[0].ID, dto.Trains[1].ID)
	}
}

func TestAddTrainRejectsDuplicateTrainID(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, err := callAddTrainForTest(t, uc, "T0", "B1", 0.0, "Up", 0.5)
	if !errors.Is(err, ErrTrainConflict) {
		t.Fatalf("expected duplicate train id error")
	}
}

func TestAddTrainRejectsOccupiedBlock(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, err := callAddTrainForTest(t, uc, "T1", "B0", 0.0, "Up", 0.5)
	if !errors.Is(err, ErrTrainConflict) {
		t.Fatalf("expected occupied block error")
	}
}

func TestAddTrainRejectsUnknownBlock(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, err := callAddTrainForTest(t, uc, "T1", "B9", 0.0, "Up", 0.5)
	if !errors.Is(err, ErrBlockNotFound) {
		t.Fatalf("expected unknown block error")
	}
}

func TestAddTrainRejectsOutOfRangeProgress(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, err := callAddTrainForTest(t, uc, "T1", "B1", 1.1, "Up", 0.5)
	if !errors.Is(err, ErrInvalidProgress) {
		t.Fatalf("expected progress validation error")
	}
}

func TestAddTrainRejectsNonPositiveSpeed(t *testing.T) {
	repo := memory.NewInMemorySimulationRepository()
	line := testLine(t)
	uc := NewUseCase(repo, &stubLineLoader{line: line})

	_, err := callAddTrainForTest(t, uc, "T1", "B1", 0.0, "Up", 0)
	if !errors.Is(err, ErrInvalidSpeed) {
		t.Fatalf("expected speed validation error")
	}
}

type stubLineLoader struct {
	line *domain.Line
	err  error
}

func (s *stubLineLoader) Load(ctx context.Context) (*domain.Line, error) {
	_ = ctx
	if s.err != nil {
		return nil, s.err
	}
	return s.line, nil
}

func testLine(t *testing.T) *domain.Line {
	t.Helper()

	s0, _ := domain.NewStationID("S0")
	s1, _ := domain.NewStationID("S1")
	s2, _ := domain.NewStationID("S2")
	b0, _ := domain.NewBlockID("B0")
	b1, _ := domain.NewBlockID("B1")

	line, err := domain.NewLine([]domain.StationID{s0, s1, s2}, []domain.BlockID{b0, b1})
	if err != nil {
		t.Fatalf("line build failed: %v", err)
	}
	return line
}

func callSetDeparturePermissionForTest(t *testing.T, uc UseCase, stationID string, allowed bool) (string, bool, error) {
	t.Helper()

	method := reflect.ValueOf(uc).MethodByName("SetDeparturePermission")
	if !method.IsValid() {
		t.Fatalf("UseCase must implement SetDeparturePermission")
	}

	methodType := method.Type()
	if methodType.NumOut() != 2 {
		t.Fatalf("SetDeparturePermission must return (dto, error), got %s", methodType.String())
	}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !methodType.Out(1).Implements(errorType) {
		t.Fatalf("SetDeparturePermission second return must be error, got %s", methodType.Out(1).String())
	}
	if methodType.NumIn() < 2 {
		t.Fatalf("SetDeparturePermission must take context and input, got %s", methodType.String())
	}
	if methodType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		t.Fatalf("SetDeparturePermission first argument must be context.Context, got %s", methodType.In(0).String())
	}

	callArgs := []reflect.Value{reflect.ValueOf(context.Background())}
	switch methodType.NumIn() {
	case 2:
		inputValue := reflect.New(methodType.In(1)).Elem()
		setStringField(t, inputValue, []string{"StationID", "StationId"}, stationID)
		setBoolField(t, inputValue, "Allowed", allowed)
		callArgs = append(callArgs, inputValue)
	case 3:
		if methodType.In(1).Kind() != reflect.String || methodType.In(2).Kind() != reflect.Bool {
			t.Fatalf("SetDeparturePermission (ctx, stationID, allowed) signature mismatch: %s", methodType.String())
		}
		callArgs = append(callArgs, reflect.ValueOf(stationID), reflect.ValueOf(allowed))
	default:
		t.Fatalf("unsupported SetDeparturePermission signature: %s", methodType.String())
	}

	results := method.Call(callArgs)
	if !results[1].IsNil() {
		callErr, ok := results[1].Interface().(error)
		if !ok {
			t.Fatalf("SetDeparturePermission returned non-error value: %T", results[1].Interface())
		}
		return "", false, callErr
	}

	dto := dereferenceValue(t, results[0])
	dtoStationID := getStringField(t, dto, []string{"StationID", "StationId"})
	dtoAllowed := getBoolField(t, dto, "Allowed")
	return dtoStationID, dtoAllowed, nil
}

func callAddTrainForTest(t *testing.T, uc UseCase, trainID string, blockID string, progress float64, direction string, speed float64) (SimulationDTO, error) {
	t.Helper()

	method := reflect.ValueOf(uc).MethodByName("AddTrain")
	if !method.IsValid() {
		t.Fatalf("UseCase must implement AddTrain")
	}

	methodType := method.Type()
	if methodType.NumOut() != 2 {
		t.Fatalf("AddTrain must return (dto, error), got %s", methodType.String())
	}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !methodType.Out(1).Implements(errorType) {
		t.Fatalf("AddTrain second return must be error, got %s", methodType.Out(1).String())
	}
	if methodType.NumIn() < 2 {
		t.Fatalf("AddTrain must take context and input, got %s", methodType.String())
	}
	if methodType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		t.Fatalf("AddTrain first argument must be context.Context, got %s", methodType.In(0).String())
	}

	callArgs := []reflect.Value{reflect.ValueOf(context.Background())}
	switch methodType.NumIn() {
	case 2:
		inputValue := reflect.New(methodType.In(1)).Elem()
		setStringField(t, inputValue, []string{"TrainID", "TrainId"}, trainID)
		setStringField(t, inputValue, []string{"BlockID", "BlockId"}, blockID)
		setFloatField(t, inputValue, []string{"Progress"}, progress)
		setDirectionField(t, inputValue, direction)
		setFloatField(t, inputValue, []string{"Speed"}, speed)
		callArgs = append(callArgs, inputValue)
	case 6:
		if methodType.In(1).Kind() != reflect.String || methodType.In(2).Kind() != reflect.String || methodType.In(3).Kind() != reflect.Float64 || methodType.In(5).Kind() != reflect.Float64 {
			t.Fatalf("AddTrain signature mismatch: %s", methodType.String())
		}
		callArgs = append(callArgs, reflect.ValueOf(trainID), reflect.ValueOf(blockID), reflect.ValueOf(progress))
		if methodType.In(4).Kind() == reflect.String {
			callArgs = append(callArgs, reflect.ValueOf(direction))
		} else if methodType.In(4).Kind() == reflect.Bool {
			callArgs = append(callArgs, reflect.ValueOf(strings.EqualFold(direction, "up")))
		} else {
			t.Fatalf("AddTrain direction argument must be string or bool, got %s", methodType.In(4).String())
		}
		callArgs = append(callArgs, reflect.ValueOf(speed))
	default:
		t.Fatalf("unsupported AddTrain signature: %s", methodType.String())
	}

	results := method.Call(callArgs)
	if !results[1].IsNil() {
		callErr, ok := results[1].Interface().(error)
		if !ok {
			t.Fatalf("AddTrain returned non-error value: %T", results[1].Interface())
		}
		return SimulationDTO{}, callErr
	}

	dtoValue := dereferenceValue(t, results[0])
	dto, ok := dtoValue.Interface().(SimulationDTO)
	if !ok {
		t.Fatalf("AddTrain first return must be SimulationDTO-compatible, got %s", dtoValue.Type().String())
	}
	return dto, nil
}

func setStringField(t *testing.T, v reflect.Value, names []string, value string) {
	t.Helper()

	target := dereferenceValue(t, v)
	for _, name := range names {
		field := target.FieldByName(name)
		if field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
			field.SetString(value)
			return
		}
	}
	t.Fatalf("missing settable string field in %s (candidates: %v)", target.Type().String(), names)
}

func setBoolField(t *testing.T, v reflect.Value, name string, value bool) {
	t.Helper()

	target := dereferenceValue(t, v)
	field := target.FieldByName(name)
	if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.Bool {
		t.Fatalf("missing settable bool field %s in %s", name, target.Type().String())
	}
	field.SetBool(value)
}

func setFloatField(t *testing.T, v reflect.Value, names []string, value float64) {
	t.Helper()

	target := dereferenceValue(t, v)
	for _, name := range names {
		field := target.FieldByName(name)
		if field.IsValid() && field.CanSet() && field.Kind() == reflect.Float64 {
			field.SetFloat(value)
			return
		}
	}
	t.Fatalf("missing settable float64 field in %s (candidates: %v)", target.Type().String(), names)
}

func setDirectionField(t *testing.T, v reflect.Value, direction string) {
	t.Helper()

	target := dereferenceValue(t, v)
	stringField := target.FieldByName("Direction")
	if stringField.IsValid() && stringField.CanSet() && stringField.Kind() == reflect.String {
		stringField.SetString(direction)
		return
	}

	boolField := target.FieldByName("Forward")
	if boolField.IsValid() && boolField.CanSet() && boolField.Kind() == reflect.Bool {
		boolField.SetBool(strings.EqualFold(direction, "up"))
		return
	}

	t.Fatalf("missing settable direction field (Direction|string or Forward|bool) in %s", target.Type().String())
}

func getStringField(t *testing.T, v reflect.Value, names []string) string {
	t.Helper()

	target := dereferenceValue(t, v)
	for _, name := range names {
		field := target.FieldByName(name)
		if field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}
	t.Fatalf("missing string field in %s (candidates: %v)", target.Type().String(), names)
	return ""
}

func getBoolField(t *testing.T, v reflect.Value, name string) bool {
	t.Helper()

	target := dereferenceValue(t, v)
	field := target.FieldByName(name)
	if !field.IsValid() || field.Kind() != reflect.Bool {
		t.Fatalf("missing bool field %s in %s", name, target.Type().String())
	}
	return field.Bool()
}

func dereferenceValue(t *testing.T, v reflect.Value) reflect.Value {
	t.Helper()

	current := v
	for current.Kind() == reflect.Ptr {
		if current.IsNil() {
			t.Fatalf("unexpected nil pointer of type %s", current.Type().String())
		}
		current = current.Elem()
	}
	return current
}
