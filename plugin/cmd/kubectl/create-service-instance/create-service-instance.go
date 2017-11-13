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

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/plugin/cmd/kubectl/utils"
)

const usage = `Usage:
  kubectl plugin create-service-instance SERVICE_CLASS_NAME PLAN_NAME INSTANCE_NAME`

func main() {
	if len(os.Args) != 4 {
		utils.Exit1(usage)
	}

	scClient, config := utils.NewClient()

	instance := v1beta1.ServiceInstance{}
	instance.Kind = "Instance"
	instance.Name = os.Args[3]
	instance.Namespace = utils.Namespace()
	instance.Spec.ClusterServicePlanRef.Name = os.Args[2]
	instance.Spec.ClusterServiceClassRef.Name = os.Args[1]

	utils.CheckNamespaceExists(instance.Namespace, config)
	utils.Ok()

	fmt.Printf("Creating service instance %s in Namespace %s...\n", utils.Entity(instance.Name), utils.Entity(instance.Namespace))
	resp, err := scClient.ServicecatalogV1beta1().ServiceInstances(instance.Namespace).Create(&instance)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Failed to create service instance (%s)", err))
	}
	utils.Ok()

	table := utils.NewTable("INSTANCE NAME", "NAMESPACE", "PLAN NAME", "SERVICE CLASS NAME")
	table.AddRow(resp.Name, resp.Namespace, resp.Spec.ClusterServicePlanRef.Name, resp.Spec.ClusterServiceClassRef.Name)
	err = table.Print()
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
	}
}
