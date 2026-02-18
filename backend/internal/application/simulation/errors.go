package simulation

import "errors"

var (
	ErrInvalidTickDelta = errors.New("invalid tick delta")
	ErrInvalidStationID = errors.New("invalid station id")
	ErrStationNotFound  = errors.New("station not found")
	ErrInvalidTrainID   = errors.New("invalid train id")
	ErrInvalidBlockID   = errors.New("invalid block id")
	ErrInvalidProgress  = errors.New("invalid progress")
	ErrInvalidDirection = errors.New("invalid direction")
	ErrInvalidSpeed     = errors.New("invalid speed")
	ErrBlockNotFound    = errors.New("block not found")
	ErrTrainConflict    = errors.New("train conflict")
)
