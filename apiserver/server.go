package main

import (
	"fmt"

	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/genericapiserver/options"
)

func main() {
	// options
	s := options.NewServerRunOptions()
	s.SecurePort = 0 // don't try to find certificates
	// genericapiserver.DefaultAndValidateRunOptions(s)

	// config
	config := genericapiserver.NewConfig()
	config = config.ApplyOptions(s)
	completedconfig := config.Complete()

	// make the server
	server, _ := completedconfig.New()
	preparedserver := server.PrepareRun()
	stop := make(chan struct{})
	fmt.Println("run the server")
	preparedserver.Run(stop)
}
