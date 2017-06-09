package server

import (
	"net/http/httptest"

	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/openservicebroker"
)

// NewCreateFunc creates a new brokerapi.CreateFunc according to a broker server running
// in srv
func NewCreateFunc(srv *httptest.Server, user, pass string) brokerapi.CreateFunc {
	// type CreateFunc func(name, url, username, password string) BrokerClient
	return brokerapi.CreateFunc(func(name, url, username, password string) brokerapi.BrokerClient {
		return openservicebroker.NewClient("testclient", srv.URL, user, pass)
	})
}
