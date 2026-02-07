package simulation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	domain "github.com/right1121/railway-control-center-simulator/internal/domain/simulation"
)

type LineLoader interface {
	Load(ctx context.Context) (*domain.Line, error)
}

type UseCase interface {
	GetSimulation(ctx context.Context) (SimulationDTO, error)
	Tick(ctx context.Context, input TickInput) (SimulationDTO, error)
}

type TickInput struct {
	DeltaMillis int64
}

type service struct {
	repo       domain.Repository
	lineLoader LineLoader
	mu         sync.Mutex
}

func NewUseCase(repo domain.Repository, lineLoader LineLoader) UseCase {
	return &service{
		repo:       repo,
		lineLoader: lineLoader,
	}
}

func (s *service) GetSimulation(ctx context.Context) (SimulationDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.ensureState(ctx)
	if err != nil {
		return SimulationDTO{}, err
	}
	return toSimulationDTO(state), nil
}

func (s *service) Tick(ctx context.Context, input TickInput) (SimulationDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delta, err := newTickDelta(input.DeltaMillis)
	if err != nil {
		return SimulationDTO{}, err
	}

	state, err := s.ensureState(ctx)
	if err != nil {
		return SimulationDTO{}, err
	}

	if err := state.Tick(delta); err != nil {
		return SimulationDTO{}, err
	}
	if err := s.repo.Save(ctx, state); err != nil {
		return SimulationDTO{}, err
	}

	return toSimulationDTO(state), nil
}

func (s *service) ensureState(ctx context.Context) (*domain.SimulationState, error) {
	state, err := s.repo.Get(ctx)
	if err == nil {
		return state, nil
	}
	if !errors.Is(err, domain.ErrSimulationNotFound) {
		return nil, err
	}

	line, err := s.lineLoader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("line load failed: %w", err)
	}
	state, err = domain.NewSimulationState(line)
	if err != nil {
		return nil, err
	}

	initialTrainID, err := domain.NewTrainID("T0")
	if err != nil {
		return nil, err
	}
	initialBlock, ok := line.BlockAt(0)
	if !ok {
		return nil, domain.ErrLineHasNoBlocks
	}
	initialProgress, err := domain.NewBlockProgress(0)
	if err != nil {
		return nil, err
	}

	initialTrain, err := domain.NewTrain(
		initialTrainID,
		initialBlock,
		initialProgress,
		true,
		0.5,
	)
	if err != nil {
		return nil, err
	}
	if err := state.AddTrain(initialTrain); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, state); err != nil {
		if errors.Is(err, domain.ErrSimulationAlreadyExists) {
			return s.repo.Get(ctx)
		}
		return nil, err
	}

	return state, nil
}

func newTickDelta(deltaMillis int64) (domain.TickDelta, error) {
	if deltaMillis <= 0 {
		return domain.TickDelta{}, ErrInvalidTickDelta
	}

	const maxInt64 = int64(^uint64(0) >> 1)
	const maxDeltaMillis = maxInt64 / int64(time.Millisecond)
	if deltaMillis > maxDeltaMillis {
		return domain.TickDelta{}, ErrInvalidTickDelta
	}

	delta, err := domain.NewTickDelta(time.Duration(deltaMillis) * time.Millisecond)
	if err != nil {
		return domain.TickDelta{}, fmt.Errorf("%w: %v", ErrInvalidTickDelta, err)
	}
	return delta, nil
}
