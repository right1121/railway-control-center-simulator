package memory

import (
	"context"
	"sync"

	domain "github.com/right1121/railway-control-center-simulator/internal/domain/simulation"
)

type InMemorySimulationRepository struct {
	mu    sync.Mutex
	state *domain.SimulationState
}

func NewInMemorySimulationRepository() domain.Repository {
	return &InMemorySimulationRepository{}
}

func (r *InMemorySimulationRepository) Get(ctx context.Context) (*domain.SimulationState, error) {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state == nil {
		return nil, domain.ErrSimulationNotFound
	}
	return r.state, nil
}

func (r *InMemorySimulationRepository) Create(ctx context.Context, state *domain.SimulationState) error {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != nil {
		return domain.ErrSimulationAlreadyExists
	}
	r.state = state
	return nil
}

func (r *InMemorySimulationRepository) Save(ctx context.Context, state *domain.SimulationState) error {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state == nil {
		return domain.ErrSimulationNotFound
	}
	r.state = state
	return nil
}
