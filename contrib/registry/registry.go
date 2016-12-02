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

	"github.com/kubernetes-incubator/service-catalog/contrib/registry/server"
	"github.com/kubernetes-incubator/service-catalog/pkg"
)

type registryOptions struct {
	ConfigPath string
	Port       int
	DefFile    string
	Version    bool
}

var options registryOptions

func init() {
	flag.StringVar(&options.ConfigPath, "c", ".", "use '-c' option to specify the config file path")
	flag.IntVar(&options.Port, "port", 8001, "use '--port' option to specify the port for registry to listen on")
	flag.StringVar(&options.DefFile, "definitions", "", "use '--definitions' option to specify a JSON definitions file to bootstrap from")

	flag.Parse()
}

func main() {
	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), pkg.VERSION)
		return
	}

	c, err := server.CreateController(server.CreateInMemoryStorage(), options.DefFile)
	if err != nil {
		panic(fmt.Sprintf("Error creating registry server: [%v]", err))
	}
	server.Start(options.Port, c)
}
