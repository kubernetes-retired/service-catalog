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

// The controller is responsible for running control loops that reconcile
// the state of service catalog API resources with service brokers, service
// classes, service instances, and service bindings.

package app

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/cmd/controller/app/options"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/watch"
	"k8s.io/client-go/1.5/dynamic"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/tools/clientcmd"
	"k8s.io/kubernetes/pkg/healthz"
)

// Run runs the ControllerServer.  This should never exit.
func Run(s *options.ControllerServer) error {
	// Create two kubernetes clients, one for normal resources and one for Third Party resources.
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", s.KubeconfigPath)
	if err != nil {
		glog.Errorf("Failed to create a kube config\n:%v\n", err)
		return err
	}

	k8sClient, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		glog.Errorf("Failed to create a kubernets client\n:%v\n", err)
		return err
	}

	kubeconfig.ContentConfig.GroupVersion = &unversioned.GroupVersion{Group: watch.GroupName, Version: watch.APIVersion}
	kubeconfig.APIPath = "apis"

	dynClient, err := dynamic.NewClient(kubeconfig)
	if err != nil {
		glog.Errorf("Failed to create a dynamic client\n:%v\n", err)
		return err
	}

	w, err := watch.NewWatcher(k8sClient, dynClient)
	if err != nil {
		panic(fmt.Sprintf("Failed to create a watcher: %v\n", err))
	}

	c, err := controller.New(w)
	if err != nil {
		panic(fmt.Sprintf("Error creating server [%s]...", err.Error()))
	}

	c.Run()

	mux := http.NewServeMux()
	healthz.InstallHandler(mux)
	server := &http.Server{
		Addr:    net.JoinHostPort(s.Address, strconv.Itoa(int(s.Port))),
		Handler: mux,
	}
	glog.Fatal(server.ListenAndServe())

	return nil
}
