package server

import (
	"github.com/pivotal-cf/brokerapi"
)

// BindRequest is the struct to contain details of a bind request
type BindRequest struct {
	InstanceID string
	BindingID  string
	Details    brokerapi.BindDetails
}
