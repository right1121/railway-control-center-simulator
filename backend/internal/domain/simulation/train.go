package simulation

import "math"

const boundaryEpsilon = 1e-9

type Train struct {
	id              TrainID
	blockID         BlockID
	progress        BlockProgress
	forward         bool
	speed           float64
	pendingTurnback bool
}

func NewTrain(id TrainID, blockID BlockID, progress BlockProgress, forward bool, speed float64) (*Train, error) {
	if speed <= 0 {
		return nil, ErrTrainSpeedNotPositive
	}
	return &Train{
		id:       id,
		blockID:  blockID,
		progress: progress,
		forward:  forward,
		speed:    speed,
	}, nil
}

func (t *Train) ID() TrainID {
	return t.id
}

func (t *Train) BlockID() BlockID {
	return t.blockID
}

func (t *Train) Progress() BlockProgress {
	return t.progress
}

func (t *Train) Forward() bool {
	return t.forward
}

func (t *Train) Speed() float64 {
	return t.speed
}

func (t *Train) PendingTurnback() bool {
	return t.pendingTurnback
}

func (t *Train) setProgress(v float64) error {
	if math.Abs(v) < boundaryEpsilon {
		v = 0
	}
	if math.Abs(v-1.0) < boundaryEpsilon {
		v = 1
	}
	progress, err := NewBlockProgress(v)
	if err != nil {
		return err
	}
	t.progress = progress
	return nil
}

func (t *Train) setBlockID(blockID BlockID) {
	t.blockID = blockID
}

func (t *Train) reverseDirection() {
	t.forward = !t.forward
}

func (t *Train) setPendingTurnback(v bool) {
	t.pendingTurnback = v
}
