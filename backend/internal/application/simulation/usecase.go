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
	SetDeparturePermission(ctx context.Context, input SetDeparturePermissionInput) (DeparturePermissionDTO, error)
	AddTrain(ctx context.Context, input AddTrainInput) (SimulationDTO, error)
}

type TickInput struct {
	DeltaMillis int64
}

type SetDeparturePermissionInput struct {
	StationID string
	Allowed   bool
}

type AddTrainInput struct {
	TrainID   string
	BlockID   string
	Progress  float64
	Direction string
	Speed     float64
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

func (s *service) SetDeparturePermission(ctx context.Context, input SetDeparturePermissionInput) (DeparturePermissionDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stationID, err := domain.NewStationID(input.StationID)
	if err != nil {
		return DeparturePermissionDTO{}, ErrInvalidStationID
	}

	state, err := s.ensureState(ctx)
	if err != nil {
		return DeparturePermissionDTO{}, err
	}

	if err := state.SetDeparturePermission(stationID, input.Allowed); err != nil {
		if errors.Is(err, domain.ErrStationNotFound) {
			return DeparturePermissionDTO{}, ErrStationNotFound
		}
		return DeparturePermissionDTO{}, err
	}

	allowed, err := state.DeparturePermission(stationID)
	if err != nil {
		if errors.Is(err, domain.ErrStationNotFound) {
			return DeparturePermissionDTO{}, ErrStationNotFound
		}
		return DeparturePermissionDTO{}, err
	}

	if err := s.repo.Save(ctx, state); err != nil {
		return DeparturePermissionDTO{}, err
	}

	return DeparturePermissionDTO{
		StationID: stationID.String(),
		Allowed:   allowed,
	}, nil
}

func (s *service) AddTrain(ctx context.Context, input AddTrainInput) (SimulationDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trainID, err := domain.NewTrainID(input.TrainID)
	if err != nil {
		return SimulationDTO{}, ErrInvalidTrainID
	}
	blockID, err := domain.NewBlockID(input.BlockID)
	if err != nil {
		return SimulationDTO{}, ErrInvalidBlockID
	}
	progress, err := domain.NewBlockProgress(input.Progress)
	if err != nil {
		return SimulationDTO{}, ErrInvalidProgress
	}
	forward, err := directionToForward(input.Direction)
	if err != nil {
		return SimulationDTO{}, err
	}

	train, err := domain.NewTrain(trainID, blockID, progress, forward, input.Speed)
	if err != nil {
		if errors.Is(err, domain.ErrTrainSpeedNotPositive) {
			return SimulationDTO{}, ErrInvalidSpeed
		}
		return SimulationDTO{}, err
	}

	state, err := s.ensureState(ctx)
	if err != nil {
		return SimulationDTO{}, err
	}

	if err := state.AddTrain(train); err != nil {
		switch {
		case errors.Is(err, domain.ErrBlockNotFound):
			return SimulationDTO{}, ErrBlockNotFound
		case errors.Is(err, domain.ErrTrainAlreadyExists), errors.Is(err, domain.ErrBlockOccupied):
			return SimulationDTO{}, ErrTrainConflict
		default:
			return SimulationDTO{}, err
		}
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

func directionToForward(direction string) (bool, error) {
	switch direction {
	case "Up", "up":
		return true, nil
	case "Down", "down":
		return false, nil
	default:
		return false, ErrInvalidDirection
	}
}
