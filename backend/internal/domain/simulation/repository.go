package simulation

import "context"

type Repository interface {
	Get(ctx context.Context) (*SimulationState, error)
	Create(ctx context.Context, state *SimulationState) error
	Save(ctx context.Context, state *SimulationState) error
}
