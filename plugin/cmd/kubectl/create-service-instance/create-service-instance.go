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

	"k8s.io/client-go/rest"
)

const usage = `Usage:
  kubectl plugin create-service-instance SERVICE_CLASS_NAME PLAN_NAME INSTANCE_NAME NAMESPACE`

func main() {
	svcURL := utils.SCUrlEnv()
	if svcURL == "" {
		svcURL = "192.168.99.100:30080"
	}

	if len(os.Args) != 5 {
		utils.Exit1(usage)
	}

	instance := v1alpha1.Instance{}
	instance.Kind = "Instance"
	instance.Name = os.Args[3]
	instance.Namespace = os.Args[4]
	instance.Spec.PlanName = os.Args[2]
	instance.Spec.ServiceClassName = os.Args[1]

	fmt.Printf("Looking up Namespace %s...\n", utils.Entity(instance.Namespace))
	if err := utils.CheckNamespaceExists(instance.Namespace); err != nil {
		utils.Exit1(err.Error())
	}
	utils.Ok()

	restConfig := rest.Config{
		Host:    svcURL,
		APIPath: "/apis/servicecatalog.k8s.io/v1alpha1",
	}

	svcClient, err := clientset.NewForConfig(&restConfig)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Failed to initializing client for service catalog (%s)", err))
	}

	fmt.Printf("Creating service instance %s in Namespace %s...\n", utils.Entity(instance.Name), utils.Entity(instance.Namespace))
	resp, err := svcClient.Instances(instance.Namespace).Create(&instance)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Failed to creating service instance (%s)", err))
	}
	utils.Ok()

	table := utils.NewTable("INSTANCE NAME", "NAMESPACE", "PLAN NAME", "SERVICE CLASS NAME")
	table.AddRow(resp.Name, resp.Namespace, resp.Spec.PlanName, resp.Spec.ServiceClassName)
	err = table.Print()
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
	}
}
