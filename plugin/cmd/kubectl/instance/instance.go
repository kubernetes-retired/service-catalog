/*
Copyright 2017 The Kubernetes Authors.

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

	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/client"
	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/utils"
)

const usage = `Usage:
  kubectl plugin instance SUBCOMMAND

Available subcommands:
  list
  get
`

const listUsage = `Usage:
  kubectl plugin instance list NAMESPACE
`

const getUsage = `Usage:
  kubectl plugin instance get NAMESPACE INSTANCE_NAME
`

func main() {
	if len(os.Args) < 2 {
		utils.Exit1(usage)
	}

	client, err := client.NewClient()
	if err != nil {
		utils.Exit1(fmt.Sprintf("Unable to initialize service catalog client (%s)", err))
	}
	if os.Args[1] == "list" {
		if len(os.Args) != 3 {
			utils.Exit1(listUsage)
		}
		namespace := os.Args[2]
		instances, err := client.ListInstances(namespace)
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to list instances in namespace %s (%s)", namespace, err))
		}

		table := utils.NewTable("INSTANCE NAME", "NAMESPACE", "CLASS NAME", "PLAN NAME")
		for _, v := range instances.Items {
			table.AddRow(v.Name, v.Namespace, v.Spec.ClusterServiceClassRef.Name, v.Spec.ClusterServicePlanRef.Name)
		}
		err = table.Print()
		if err != nil {
			utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
		}
	} else if os.Args[1] == "get" {
		if len(os.Args) != 4 {
			utils.Exit1(getUsage)
		}
		namespace := os.Args[2]
		instanceName := os.Args[3]
		instance, err := client.GetInstance(instanceName, namespace)
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to find instance %s in namespace %s (%s)", instanceName, namespace, err))
		}
		table := utils.NewTable("INSTANCE NAME", "NAMESPACE", "CLASS NAME", "PLAN NAME")
		table.AddRow(instance.Name, instance.Namespace, instance.Spec.ClusterServiceClassRef.Name, instance.Spec.ClusterServicePlanRef.Name)
		err = table.Print()
	} else {
		utils.Exit1(usage)
	}
}
