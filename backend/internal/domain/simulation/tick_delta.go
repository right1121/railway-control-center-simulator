package simulation

import "time"

type TickDelta struct {
	value time.Duration
}

func NewTickDelta(v time.Duration) (TickDelta, error) {
	if v <= 0 {
		return TickDelta{}, ErrTickDeltaNotPositive
	}
	return TickDelta{value: v}, nil
}

func (d TickDelta) Duration() time.Duration {
	return d.value
}
