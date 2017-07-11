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

	v1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	clientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/plugin/cmd/kubectl/utils"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

const usage = `Usage:
  kubectl plugin bind-service INSTANCE_NAME BINDING_NAME NAMESPACE`

func main() {
	svcURL := utils.SCUrlEnv()
	if svcURL == "" {
		svcURL = "192.168.99.100:30080"
	}

	if len(os.Args) != 4 {
		utils.Exit1(usage)
	}

	binding := v1alpha1.Binding{}
	binding.Kind = "binding"
	binding.Name = os.Args[2]
	binding.Namespace = os.Args[3]
	binding.Spec.InstanceRef = v1.LocalObjectReference{
		Name: os.Args[1],
	}
	binding.Spec.SecretName = os.Args[2]

	fmt.Printf("Looking up Namespace %s...\n", utils.Entity(binding.Namespace))
	if err := utils.CheckNamespaceExists(binding.Namespace); err != nil {
		utils.Exit1(err.Error())
	}
	utils.Ok()

	restConfig := rest.Config{
		Host:    svcURL,
		APIPath: "/apis/servicecatalog.k8s.io/v1alpha1",
	}

	svcClient, err := clientset.NewForConfig(&restConfig)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error initializing client for service catalog (%s)", err))
	}

	fmt.Printf("Creating binding %s to %s in Namespace %s...\n", utils.Entity(binding.Name), utils.Entity(binding.Spec.InstanceRef.Name), utils.Entity(binding.Namespace))
	resp, err := svcClient.Bindings(binding.Namespace).Create(&binding)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error binding service instance (%s)", err))
	}
	utils.Ok()

	table := utils.NewTable("BINDING NAME", "NAMESPACE", "INSTANCE NAME", "SECRET NAME")
	table.AddRow(resp.Name, resp.Namespace, resp.Spec.InstanceRef.Name, resp.Spec.SecretName)
	err = table.Print()
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
	}
}
