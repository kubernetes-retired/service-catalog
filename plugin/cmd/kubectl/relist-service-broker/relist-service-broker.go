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

	"github.com/kubernetes-incubator/service-catalog/plugin/cmd/kubectl/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const usage = `Usage:
  kubectl plugin relist-service-broker BROKER_NAME`

func main() {
	if len(os.Args) != 2 {
		utils.Exit1(usage)
	}

	scClient, _ := utils.NewClient()
	brokerName := os.Args[1]
	broker, err := scClient.ServicecatalogV1beta1().ClusterServiceBrokers().Get(brokerName, v1.GetOptions{})
	if err != nil {
		utils.Exit1(fmt.Sprintf("Unable to find broker %s (%s)", brokerName, err))
	}

	fmt.Printf("Relisting broker %s...\n", utils.Entity(broker.Name))
	broker.Spec.RelistRequests++
	resp, err := scClient.ServicecatalogV1beta1().ClusterServiceBrokers().Update(broker)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Updating broker resource (%s)", err))
	}

	utils.Ok()

	table := utils.NewTable("BROKER NAME", "NAMESPACE", "URL")
	table.AddRow(resp.Name, resp.Namespace, resp.Spec.URL)
	err = table.Print()
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
	}
}
