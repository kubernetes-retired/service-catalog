/*
Copyright 2016 The Kubernetes Authors.

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

// The apiserver is the api server and master for the service catalog.
// It is responsible for serving the service catalog management API.

package main

import (
	"os"

	"github.com/golang/glog"
	// set up logging the k8s way
	"k8s.io/kubernetes/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/util/logs"

	"github.com/kubernetes-incubator/service-catalog/cmd/apiserver/app/server"
	// install our API groups
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	"k8s.io/kubernetes/pkg/client/restclient"
)

func main() {
	logs.InitLogs()
	// make sure we print all the logs while shutting down.
	defer logs.FlushLogs()

	cfg, err := restclient.InClusterConfig()
	if err != nil {
		glog.Errorf("Failed to get kube client config (%s)", err)
		os.Exit(1)
	}
	cfg.GroupVersion = &schema.GroupVersion{}

	clIface, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("Failed to create clientset Interface (%s)", err)
		os.Exit(1)
	}

	cmd := server.NewCommandServer(os.Stdout, clIface)
	if err := cmd.Execute(); err != nil {
		glog.Errorf("server exited unexpectedly (%s)", err)
		logs.FlushLogs()
		os.Exit(1)
	}
}
