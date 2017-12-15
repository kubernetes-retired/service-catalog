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
  kubectl plugin broker SUBCOMMAND

Available subcommands:
  list
  get
`

const getUsage = `Usage:
  kubectl plugin broker get BROKERNAME
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
		brokers, err := client.ListBrokers()
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to list brokers (%s)", err))
		}

		table := utils.NewTable("BROKER NAME", "NAMESPACE", "URL")
		for _, v := range brokers.Items {
			table.AddRow(v.Name, v.Namespace, v.Spec.URL)
			err = table.Print()
		}
		if err != nil {
			utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
		}
	} else if os.Args[1] == "get" {
		if len(os.Args) != 3 {
			utils.Exit1(getUsage)
		}
		brokerName := os.Args[2]
		broker, err := client.GetBroker(brokerName)
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to find broker %s (%s)", brokerName, err))
		}
		table := utils.NewTable("BROKER NAME", "NAMESPACE", "URL")
		table.AddRow(broker.Name, broker.Namespace, broker.Spec.URL)
		err = table.Print()
	} else {
		utils.Exit1(usage)
	}
}
