package server

import (
	"github.com/pivotal-cf/brokerapi"
)

// UnbindRequest is the struct to house details of a single unbind request
type UnbindRequest struct{
	InstanceID string
	BindingID string
	Details brokerapi.UnbindDetails
}
