package simulation

import "errors"

var (
	ErrInvalidTickDelta = errors.New("invalid tick delta")
	ErrInvalidStationID = errors.New("invalid station id")
	ErrStationNotFound  = errors.New("station not found")
)
