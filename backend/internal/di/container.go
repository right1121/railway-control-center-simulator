package di

import (
	sessionapp "github.com/right1121/railway-control-center-simulator/internal/application/session"
	simulationapp "github.com/right1121/railway-control-center-simulator/internal/application/simulation"
	"github.com/right1121/railway-control-center-simulator/internal/config"
	"github.com/right1121/railway-control-center-simulator/internal/domain/session"
	"github.com/right1121/railway-control-center-simulator/internal/domain/simulation"
	lineLoader "github.com/right1121/railway-control-center-simulator/internal/infrastructure/filesystem"
	sessionRepo "github.com/right1121/railway-control-center-simulator/internal/infrastructure/memory"
)

// Container は依存の生成・保持を担当する。
// - 生成コストがあるものはシングルトンとして保持
type Container struct {
	cfg *config.Config

	Repositories Repositories

	UseCases UseCases
}

type Repositories struct {
	Session    session.Repository
	Simulation simulation.Repository
}

type UseCases struct {
	Session    sessionapp.UseCase
	Simulation simulationapp.UseCase
}

// NewContainer は DI コンテナを生成する。
// ここではまだ依存を組み立てず、遅延初期化する（起動を軽くする）。
func NewContainer(cfg *config.Config) *Container {
	session := sessionRepo.NewInMemorySessionRepository()
	simState := sessionRepo.NewInMemorySimulationRepository()
	loader := lineLoader.NewSimulationLineLoader(lineLoader.DefaultSimulationLinePath)

	repos := Repositories{
		Session:    session,
		Simulation: simState,
	}

	usecase := UseCases{
		Session:    sessionapp.NewUseCase(repos.Session),
		Simulation: simulationapp.NewUseCase(repos.Simulation, loader),
	}

	return &Container{
		cfg:          cfg,
		Repositories: repos,
		UseCases:     usecase,
	}
}
