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
  kubectl plugin plan SUBCOMMAND

Available subcommands:
  list
  get
`

const getUsage = `Usage:
  kubectl plugin plan get PLAN_NAME
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
		plans, err := client.ListPlans()
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to list plans (%s)", err))
		}

		table := utils.NewTable("PLAN NAME", "DESCRIPTION", "BROKER NAME")
		for _, v := range plans.Items {
			table.AddRow(v.Spec.ExternalName, v.Spec.Description, v.Spec.ClusterServiceBrokerName)
		}
		err = table.Print()
		if err != nil {
			utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
		}
	} else if os.Args[1] == "get" {
		if len(os.Args) != 3 {
			utils.Exit1(getUsage)
		}
		planName := os.Args[2]
		plan, err := client.GetPlan(planName)
		if err != nil {
			utils.Exit1(fmt.Sprintf("Unable to find plan %s (%s)", planName, err))
		}
		table := utils.NewTable("PLAN NAME", "DESCRIPTIONS", "BROKER NAME")
		table.AddRow(plan.Name, plan.Spec.Description, plan.Spec.ClusterServiceBrokerName)
		err = table.Print()
	} else {
		utils.Exit1(usage)
	}
}
