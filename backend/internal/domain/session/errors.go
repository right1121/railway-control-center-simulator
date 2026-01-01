package session

import "errors"

var (
	ErrSessionAlreadyExists    = errors.New("session already exists")
	ErrDispatcherAlreadyExists = errors.New("dispatcher already exists")
	ErrDispatcherNotFound      = errors.New("dispatcher not found")
)
