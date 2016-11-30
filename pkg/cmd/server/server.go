package server

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/genericapiserver"
	options "k8s.io/kubernetes/pkg/genericapiserver/options"
)

func NewCommandServer(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "run a service-catalog server",
		Run: func(c *cobra.Command, args []string) {
			runServer()
		},
	}

	// flags := cmd.Flags()

	return cmd
}

func runServer() {
	fmt.Println("set up the server")
	// options
	s := options.NewServerRunOptions()
	s.SecurePort = 0 // don't try to find certificates
	// genericapiserver.DefaultAndValidateRunOptions(s)

	// config
	config := genericapiserver.NewConfig()
	config = config.ApplyOptions(s)

	/*
		if err := genericAPIServerConfig.MaybeGenerateServingCerts(); err != nil {
			return err
		}
	*/

	completedconfig := config.Complete()

	// make the server
	server, _ := completedconfig.New()
	preparedserver := server.PrepareRun()

	// set up storage

	stop := make(chan struct{})
	fmt.Println("run the server")
	preparedserver.Run(stop)
}
