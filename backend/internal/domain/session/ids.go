package session

import (
	"errors"
	"strings"
)

// DispatcherID は管制員IDの Value Object
type DispatcherID struct {
	value string
}

func NewDispatcherID(v string) (DispatcherID, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return DispatcherID{}, errors.New("dispatcher id is empty")
	}
	return DispatcherID{value: v}, nil
}

func (id DispatcherID) String() string {
	return id.value
}

// SessionID は訓練セッションIDの Value Object
type SessionID struct {
	value string
}

func NewSessionID(v string) (SessionID, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return SessionID{}, errors.New("session id is empty")
	}
	return SessionID{value: v}, nil
}

func (id SessionID) String() string {
	return id.value
}
