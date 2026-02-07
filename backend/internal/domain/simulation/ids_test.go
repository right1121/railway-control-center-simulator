package simulation

import (
	"testing"
	"time"
)

func TestNewTrainIDEmpty(t *testing.T) {
	if _, err := NewTrainID("   "); err != ErrTrainIDEmpty {
		t.Fatalf("expected ErrTrainIDEmpty, got %v", err)
	}
}

func TestNewBlockProgressRange(t *testing.T) {
	if _, err := NewBlockProgress(-0.1); err != ErrBlockProgressOutOfRange {
		t.Fatalf("expected ErrBlockProgressOutOfRange, got %v", err)
	}
	if _, err := NewBlockProgress(1.1); err != ErrBlockProgressOutOfRange {
		t.Fatalf("expected ErrBlockProgressOutOfRange, got %v", err)
	}
}

func TestNewTickDelta(t *testing.T) {
	if _, err := NewTickDelta(0); err != ErrTickDeltaNotPositive {
		t.Fatalf("expected ErrTickDeltaNotPositive, got %v", err)
	}
	if _, err := NewTickDelta(-1 * time.Second); err != ErrTickDeltaNotPositive {
		t.Fatalf("expected ErrTickDeltaNotPositive, got %v", err)
	}
}
