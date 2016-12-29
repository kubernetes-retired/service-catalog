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

	"github.com/kubernetes-incubator/service-catalog/pkg"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/server"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/catalog/watch"
	"k8s.io/client-go/1.5/dynamic"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/tools/clientcmd"
)

var options struct {
	ConfigPath string
	Port       int
	kubeconfig string
}

func init() {
	flag.StringVar(&options.ConfigPath, "c", ".", "use '-c' option to specify the config file path")
	flag.IntVar(&options.Port, "port", 10000, "use '--port' option to specify the port for controller to listen on")
	flag.StringVar(&options.kubeconfig, "kubeconfig", "./kubeconfig", "Path to kubeconfig")

	flag.Parse()
}

func main() {
	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), pkg.VERSION)
		return
	}

	// Create two kubernetes clients, one for normal resources and one for Third Party resources.
	config, err := clientcmd.BuildConfigFromFlags("", options.kubeconfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create a kube config\n:%v\n", err))
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create a kubernets client\n:%v\n", err))
	}

	// this conversion is probably gonna be done in generated code later. for now, manual :(
	config.ContentConfig.GroupVersion = &unversioned.GroupVersion{
		Group:   v1alpha1.GroupVersion.Group,
		Version: v1alpha1.GroupVersion.Version,
	}
	config.APIPath = "apis"

	dynClient, err := dynamic.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create a dynamic client\n:%v\n", err))
	}

	w, err := watch.NewWatcher(k8sClient, dynClient)
	if err != nil {
		panic(fmt.Sprintf("Failed to create a watcher: %v\n", err))
	}

	s, err := server.CreateServer(w)
	if err != nil {
		panic(fmt.Sprintf("Error creating server [%s]...", err.Error()))
	}

	s.Start(options.Port)
}
