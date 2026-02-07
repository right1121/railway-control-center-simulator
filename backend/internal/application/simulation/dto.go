package simulation

import domain "github.com/right1121/railway-control-center-simulator/internal/domain/simulation"

type SimulationDTO struct {
	SimTimeMillis int64      `json:"simTimeMillis"`
	Line          LineDTO    `json:"line"`
	Trains        []TrainDTO `json:"trains"`
}

type LineDTO struct {
	Stations []string `json:"stations"`
	Blocks   []string `json:"blocks"`
}

type TrainDTO struct {
	ID              string  `json:"id"`
	BlockID         string  `json:"blockId"`
	Progress        float64 `json:"progress"`
	Forward         bool    `json:"forward"`
	Speed           float64 `json:"speed"`
	PendingTurnback bool    `json:"pendingTurnback"`
}

func toSimulationDTO(state *domain.SimulationState) SimulationDTO {
	line := state.Line()
	stations := line.Stations()
	blocks := line.Blocks()
	trains := state.Trains()

	stationIDs := make([]string, 0, len(stations))
	for _, station := range stations {
		stationIDs = append(stationIDs, station.String())
	}

	blockIDs := make([]string, 0, len(blocks))
	for _, block := range blocks {
		blockIDs = append(blockIDs, block.String())
	}

	trainDTOs := make([]TrainDTO, 0, len(trains))
	for _, train := range trains {
		trainDTOs = append(trainDTOs, TrainDTO{
			ID:              train.ID().String(),
			BlockID:         train.BlockID().String(),
			Progress:        train.Progress().Float64(),
			Forward:         train.Forward(),
			Speed:           train.Speed(),
			PendingTurnback: train.PendingTurnback(),
		})
	}

	return SimulationDTO{
		SimTimeMillis: state.SimTime().Millis(),
		Line: LineDTO{
			Stations: stationIDs,
			Blocks:   blockIDs,
		},
		Trains: trainDTOs,
	}
}
