package session

import "errors"

var (
	ErrDispatcherAlreadyExists = errors.New("dispatcher already exists")
	ErrDispatcherNotFound      = errors.New("dispatcher not found")
)
