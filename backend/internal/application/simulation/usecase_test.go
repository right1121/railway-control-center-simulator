package simulation

import (
	"context"
	"errors"
	"math"
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
