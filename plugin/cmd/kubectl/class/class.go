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
  kubectl plugin class SUBCOMMAND

Available subcommands:
  list
  get
`

const getUsage = `Usage:
  kubectl plugin class get CLASSNAME
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
		classes, err := client.ListClasses()
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to list classes (%s)", err))
		}

		table := utils.NewTable("CLASS NAME", "NAMESPACE", "BROKER NAME")
		for _, v := range classes.Items {
			table.AddRow(v.Name, v.Namespace, v.Spec.ClusterServiceBrokerName)
			err = table.Print()
		}
		if err != nil {
			utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
		}
	} else if os.Args[1] == "get" {
		if len(os.Args) != 3 {
			utils.Exit1(getUsage)
		}
		className := os.Args[2]
		class, err := client.GetClass(className)
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to find class %s (%s)", className, err))
		}
		table := utils.NewTable("CLASS NAME", "NAMESPACE", "BROKER NAME")
		table.AddRow(class.Name, class.Namespace, class.Spec.ClusterServiceBrokerName)
		err = table.Print()
	} else {
		utils.Exit1(usage)
	}
}
