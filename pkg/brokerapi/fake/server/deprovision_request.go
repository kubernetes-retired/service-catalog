package server

import (
	"github.com/pivotal-cf/brokerapi"
)

// DeprovisionRequest is the struct to contain details of a single deprovision request
type DeprovisionRequest struct {
	InstanceID string
	Details    brokerapi.DeprovisionDetails
}
