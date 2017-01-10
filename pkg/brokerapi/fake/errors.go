package fake

import (
	"errors"
)

var (
	ErrInstanceNotFound      = errors.New("instance not found")
	ErrInstanceAlreadyExists = errors.New("instance already exists")
	ErrBindingNotFound       = errors.New("binding not found")
	ErrBindingAlreadyExists  = errors.New("binding already exists")
)
