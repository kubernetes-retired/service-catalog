package brokerapi

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

// CreateFunc is a callback function that the controller uses to create a new broker client
// from a given broker object
type CreateFunc func(*servicecatalog.Broker) BrokerClient
