package server

import (
	"github.com/pivotal-cf/brokerapi"
)

// ProvisionRequest is the struct to house details of a single provision request
type ProvisionRequest struct {
	InstanceID   string
	Details      brokerapi.ProvisionDetails
	AsyncAllowed bool
}
