package main

import (
	"fmt"
	"os"

	v1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
	clientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/plugin/cmd/kubectl/utils"

	"k8s.io/client-go/rest"
)

const USAGE = `Usage:
  kubectl plugin create-service-broker BROKER_NAME BROKER_URL`

func main() {
	svcURL := utils.SCUrlEnv()
	if svcURL == "" {
		svcURL = "192.168.99.100:30080"
	}

	if len(os.Args) != 3 {
		utils.Exit1(USAGE)
	}

	broker := v1alpha1.Broker{}
	broker.Kind = "Broker"
	broker.Name = os.Args[1]
	broker.Spec.URL = os.Args[2]

	restConfig := rest.Config{
		Host:    svcURL,
		APIPath: "/apis/servicecatalog.k8s.io/v1alpha1",
	}

	svcClient, err := clientset.NewForConfig(&restConfig)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Initializing client for service catalog (%s)", err))
	}

	fmt.Printf("Creating broker %s...\n", utils.Entity(broker.Name))
	resp, err := svcClient.Brokers().Create(&broker)
	if err != nil {
		utils.Exit1(fmt.Sprintf("Creating broker resource (%s)", err))
	}

	utils.Ok()

	table := utils.NewTable("BROKER NAME", "NAMESPACE", "URL")
	table.AddRow(resp.Name, resp.Namespace, resp.Spec.URL)
	err = table.Print()
	if err != nil {
		utils.Exit1(fmt.Sprintf("Error printing result (%s)", err))
	}
}
