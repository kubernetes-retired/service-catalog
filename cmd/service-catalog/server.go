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

package main

import (
	"fmt"
	"os"
	// "runtime"

	// commented out until I know what this does
	// "k8s.io/kubernetes/pkg/util/logs"

	"github.com/kubernetes-incubator/service-catalog/pkg/cmd/server"
	// commented out until I know what this does
	// install all APIs
	// _ "github.com/openshift/kube-aggregator/pkg/apis/apifederation/install"
	// _ "k8s.io/kubernetes/pkg/api/install"
)

func main() {
	// commented out until I know what this does
	// logs.InitLogs()
	// defer logs

	cmd := server.NewCommandServer(os.Stdout)
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
