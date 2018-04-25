/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/apiserver/pkg/server/healthz"
)

// ServeHTTP starts a new Http Server thread for /metrics and health probing
func ServeHTTP(healthcheckOptions *HealthCheckServer) error {

	// Initialize SSL/TLS configuration.  Creats a self signed certificate and key if necessary
	if err := healthcheckOptions.SecureServingOptions.MaybeDefaultWithSelfSignedCerts("" /*AdvertiseAddress*/, nil /*alternateDNS*/, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return fmt.Errorf("failed to establish SecureServingOptions %v", err)
	}

	glog.V(3).Infof("Starting http server and mux on port %v", healthcheckOptions.SecureServingOptions.BindPort)

	go func() {
		mux := http.NewServeMux()

		RegisterMetricsAndInstallHandler(mux)
		healthz.InstallHandler(mux, healthz.PingHealthz)

		server := &http.Server{
			Addr: net.JoinHostPort(healthcheckOptions.SecureServingOptions.BindAddress.String(),
				strconv.Itoa(healthcheckOptions.SecureServingOptions.BindPort)),
			Handler: mux,
		}
		glog.Fatal(server.ListenAndServeTLS(healthcheckOptions.SecureServingOptions.ServerCert.CertKey.CertFile,
			healthcheckOptions.SecureServingOptions.ServerCert.CertKey.KeyFile))
	}()
	return nil
}
