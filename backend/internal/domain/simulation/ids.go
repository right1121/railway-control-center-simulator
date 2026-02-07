package simulation

import (
	"strings"
	"time"
)

type TrainID struct{ value string }

func NewTrainID(v string) (TrainID, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return TrainID{}, ErrTrainIDEmpty
	}
	return TrainID{value: v}, nil
}

func (id TrainID) String() string {
	return id.value
}

type BlockID struct{ value string }
type StationID struct{ value string }

func NewBlockID(v string) (BlockID, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return BlockID{}, ErrBlockIDEmpty
	}
	return BlockID{value: v}, nil
}

func (id BlockID) String() string {
	return id.value
}

func NewStationID(v string) (StationID, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return StationID{}, ErrStationIDEmpty
	}
	return StationID{value: v}, nil
}

func (id StationID) String() string {
	return id.value
}

type SimTime struct{ millis int64 }

func (t SimTime) Add(dt time.Duration) SimTime {
	return SimTime{millis: t.millis + dt.Milliseconds()}
}

func (t SimTime) Millis() int64 {
	return t.millis
}

type BlockProgress struct{ value float64 } // 0..1

func NewBlockProgress(v float64) (BlockProgress, error) {
	if v < 0.0 || v > 1.0 {
		return BlockProgress{}, ErrBlockProgressOutOfRange
	}
	return BlockProgress{value: v}, nil
}

func (p BlockProgress) Float64() float64 {
	return p.value
}
