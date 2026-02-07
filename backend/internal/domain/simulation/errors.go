package simulation

import "errors"

var (
	ErrTrainIDEmpty               = errors.New("train id is empty")
	ErrBlockIDEmpty               = errors.New("block id is empty")
	ErrStationIDEmpty             = errors.New("station id is empty")
	ErrBlockProgressOutOfRange    = errors.New("block progress must be in range [0,1]")
	ErrTickDeltaNotPositive       = errors.New("tick delta must be greater than zero")
	ErrTrainSpeedNotPositive      = errors.New("train speed must be greater than zero")
	ErrLineHasNoBlocks            = errors.New("line must have at least one block")
	ErrLineStationsBlocksMismatch = errors.New("line must have blocks+1 stations")
	ErrLineDuplicateStationID     = errors.New("line has duplicate station id")
	ErrLineDuplicateBlockID       = errors.New("line has duplicate block id")
	ErrLineConnectivityInvalid    = errors.New("line connectivity is invalid")
	ErrBlockNotFound              = errors.New("block not found")
	ErrTrainAlreadyExists         = errors.New("train already exists")
	ErrTrainNotFound              = errors.New("train not found")
	ErrBlockOccupied              = errors.New("block is occupied")
	ErrSimulationNotFound         = errors.New("simulation not found")
	ErrSimulationAlreadyExists    = errors.New("simulation already exists")
)
