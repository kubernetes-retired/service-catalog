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
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/kubernetes-incubator/service-catalog/broker/k8s/controller"
	"github.com/kubernetes-incubator/service-catalog/broker/server"
	"github.com/kubernetes-incubator/service-catalog/pkg"
)

type k8sBrokerOptions struct {
	ConfigPath   string
	Port         int
	RegistryHost string
	RegistryPort int
	HelmBinary   string
	TillerHost   string
}

var options k8sBrokerOptions

func init() {
	flag.StringVar(&options.ConfigPath, "c", ".", "use '-c' option to specify the config file path")
	flag.IntVar(&options.Port, "port", 8000, "use '--port' option to specify the port for broker to listen on")
	flag.StringVar(&options.RegistryHost, "registry_host", "localhost", "use '--registry_host' option to specify the hostname for registry")
	flag.IntVar(&options.RegistryPort, "registry_port", 8001, "use '--registry_port' option to specify the port for registry")
	flag.StringVar(&options.HelmBinary, "helm_binary", "", "full path to helm binary")
	flag.StringVar(&options.TillerHost, "tiller_host", "", "tiller host spec")
	flag.Parse()
}

func main() {
	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), pkg.VERSION)
		return
	}

	if len(options.HelmBinary) == 0 {
		panic("Need helm_binary specified")
	}
	r := controller.NewHelmReifier(options.HelmBinary, options.TillerHost)
	c, err := controller.CreateController(options.RegistryHost, options.RegistryPort, r)
	if err != nil {
		panic(fmt.Sprintf("Error creating controller [%s]...", err.Error()))
	}

	server.Start(options.Port, c)
}
