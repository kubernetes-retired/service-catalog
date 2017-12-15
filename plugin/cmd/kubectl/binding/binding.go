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

	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/client"
	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/utils"
)

const usage = `Usage:
  kubectl plugin binding SUBCOMMAND

Available subcommands:
  list
  get
`

const listUsage = `Usage:
  kubectl plugin binding list NAMESPACE
`

const getUsage = `Usage:
  kubectl plugin binding get NAMESPACE INSTANCENAME
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
		bindings, err := client.ListBindings(namespace)
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to list bindings in namespace %s (%s)", namespace, err))
		}

		table := utils.NewTable("BINDING NAME", "NAMESPACE", "INSTANCE NAME")
		for _, v := range bindings.Items {
			table.AddRow(v.Name, v.Namespace, v.Spec.ServiceInstanceRef.Name)
			err = table.Print()
		}
		if err != nil {
			utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
		}
	} else if os.Args[1] == "get" {
		if len(os.Args) != 4 {
			utils.Exit1(getUsage)
		}
		namespace := os.Args[2]
		bindingName := os.Args[3]
		binding, err := client.GetBinding(bindingName, namespace)
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to find binding %s in namespae %s (%s)", bindingName, namespace, err))
		}
		table := utils.NewTable("BINDINGNAME", "NAMESPACE", "INSTANCE NAME")
		table.AddRow(binding.Name, binding.Namespace, binding.Spec.ServiceInstanceRef.Name)
		err = table.Print()
	} else {
		utils.Exit1(usage)
	}
}
