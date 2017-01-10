package fake

import (
	"errors"
)

var (
	// ErrInstanceNotFound is returned by the fake client when an instance was expected in its
	// internal storage but it wasn't found
	ErrInstanceNotFound = errors.New("instance not found")
	// ErrInstanceAlreadyExists is returned by the fake client when an instance was found in
	// internal storage but it wasn't expected
	ErrInstanceAlreadyExists = errors.New("instance already exists")
	// ErrBindingNotFound is returned by the fake client when a binding was expected in internal
	// storage but it wasn't found
	ErrBindingNotFound = errors.New("binding not found")
	// ErrBindingAlreadyExists is returned by the fake client when a binding was found in internal
	// storage but it wasn't expected
	ErrBindingAlreadyExists = errors.New("binding already exists")
)
