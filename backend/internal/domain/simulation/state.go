package simulation

import (
	"math"
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
	if train == nil {
		return ErrTrainNotFound
	}

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
	if dt.Duration() <= 0 {
		return ErrTickDeltaNotPositive
	}

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

func (s *SimulationState) IsAtBoundary(trainID TrainID) bool {
	train, ok := s.trains[trainID.String()]
	if !ok {
		return false
	}
	return isBoundaryProgress(train.Progress().Float64())
}

func (s *SimulationState) TrainStation(trainID TrainID) (StationID, bool) {
	return s.TrainStationAtBoundary(trainID)
}

func (s *SimulationState) TrainStationAtBoundary(trainID TrainID) (StationID, bool) {
	train, ok := s.trains[trainID.String()]
	if !ok {
		return StationID{}, false
	}

	progress := train.Progress().Float64()
	if train.Forward() {
		if !isProgressOne(progress) {
			return StationID{}, false
		}
		return s.line.ToStation(train.BlockID())
	}
	if !isProgressZero(progress) {
		return StationID{}, false
	}
	return s.line.FromStation(train.BlockID())
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

func isBoundaryProgress(progress float64) bool {
	return isProgressZero(progress) || isProgressOne(progress)
}

func isProgressZero(progress float64) bool {
	return math.Abs(progress) < boundaryEpsilon
}

func isProgressOne(progress float64) bool {
	return math.Abs(progress-1) < boundaryEpsilon
}
