package server

import (
	"github.com/pivotal-cf/brokerapi"
)

// UpdateRequest is the struct that contains details of a single update request
type UpdateRequest struct {
	InstanceID   string
	Details      brokerapi.UpdateDetails
	AsyncAllowed bool
}
