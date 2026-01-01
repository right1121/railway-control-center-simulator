package session

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// DispatcherName は管制員の表示名を表す Value Object
type DispatcherName struct {
	value string
}

func NewDispatcherName(name string) (DispatcherName, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return DispatcherName{}, errors.New("dispatcher name is required")
	}

	if utf8.RuneCountInString(name) > 32 {
		return DispatcherName{}, errors.New("dispatcher name is too long")
	}

	return DispatcherName{value: name}, nil
}

func (n DispatcherName) String() string {
	return n.value
}
