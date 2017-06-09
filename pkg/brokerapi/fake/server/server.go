package server

import (
	"net/http/httptest"

	"github.com/pivotal-cf/brokerapi"
)

// Run runs a new test server from the given broker handler and auth credentials
func Run(hdl *Handler, username, password string) *httptest.Server {
	httpHandler := brokerapi.New(hdl, logger, brokerapi.BrokerCredentials{
		Username: username,
		Password: password,
	})

	srv := httptest.NewServer(httpHandler)
	return srv
}
