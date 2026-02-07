package simulation

import (
	"sort"
)

type SimulationState struct {
	line     *Line
	simTime  SimTime
	trains   map[string]*Train
	occupied map[string]TrainID
}

func NewSimulationState(line *Line) (*SimulationState, error) {
	if line == nil {
		return nil, ErrLineHasNoBlocks
	}
	return &SimulationState{
		line:     line,
		trains:   make(map[string]*Train),
		occupied: make(map[string]TrainID),
	}, nil
}

func (s *SimulationState) Line() *Line {
	return s.line
}

func (s *SimulationState) SimTime() SimTime {
	return s.simTime
}

func (s *SimulationState) Trains() []Train {
	keys := s.sortedTrainKeys()
	out := make([]Train, 0, len(keys))
	for _, key := range keys {
		out = append(out, *s.trains[key])
	}
	return out
}

func (s *SimulationState) AddTrain(train *Train) error {
	trainKey := train.ID().String()
	if _, exists := s.trains[trainKey]; exists {
		return ErrTrainAlreadyExists
	}
	if !s.line.HasBlock(train.BlockID()) {
		return ErrBlockNotFound
	}

	blockKey := train.BlockID().String()
	if _, occupied := s.occupied[blockKey]; occupied {
		return ErrBlockOccupied
	}

	s.trains[trainKey] = train
	s.occupied[blockKey] = train.ID()
	return nil
}

func (s *SimulationState) Tick(dt TickDelta) error {
	s.simTime = s.simTime.Add(dt.Duration())

	keys := s.sortedTrainKeys()
	for _, key := range keys {
		train := s.trains[key]
		if train.PendingTurnback() {
			train.reverseDirection()
			train.setPendingTurnback(false)
		}
	}

	for _, key := range keys {
		train := s.trains[key]
		distance := train.Speed() * dt.Duration().Seconds()

		for distance > 0 {
			progress := train.Progress().Float64()
			remaining := progress
			if train.Forward() {
				remaining = 1.0 - progress
			}
			if remaining < boundaryEpsilon {
				remaining = 0
			}

			if distance+boundaryEpsilon < remaining {
				nextProgress := progress - distance
				if train.Forward() {
					nextProgress = progress + distance
				}
				if err := train.setProgress(nextProgress); err != nil {
					return err
				}
				distance = 0
				continue
			}

			if train.Forward() {
				if err := train.setProgress(1); err != nil {
					return err
				}
			} else {
				if err := train.setProgress(0); err != nil {
					return err
				}
			}
			distance -= remaining
			if distance < boundaryEpsilon {
				distance = 0
			}

			nextBlock, exists, err := s.line.NextBlock(train.BlockID(), train.Forward())
			if err != nil {
				return err
			}
			if !exists {
				train.setPendingTurnback(true)
				break
			}

			if occupiedBy, occupied := s.occupied[nextBlock.String()]; occupied && occupiedBy.String() != train.ID().String() {
				break
			}

			delete(s.occupied, train.BlockID().String())
			train.setBlockID(nextBlock)
			s.occupied[nextBlock.String()] = train.ID()

			if train.Forward() {
				if err := train.setProgress(0); err != nil {
					return err
				}
			} else {
				if err := train.setProgress(1); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *SimulationState) sortedTrainKeys() []string {
	keys := make([]string, 0, len(s.trains))
	for key := range s.trains {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (l *Line) HasBlock(id BlockID) bool {
	_, ok := l.blockIndex[id.String()]
	return ok
}
