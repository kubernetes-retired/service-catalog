package server

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/genericapiserver"
	genericserveroptions "k8s.io/kubernetes/pkg/genericapiserver/options"
)

// ServerOptions contains the aggregation of configuration structs for the service-catalog server
type ServerOptions struct {
	// for etcd storage ?
	EtcdOptions *genericserveroptions.EtcdOptions
	// for the http stuff ?
	SecureServingOptions *genericserveroptions.SecureServingOptions
}

func NewCommandServer(out io.Writer) *cobra.Command {
	options := &ServerOptions{
		EtcdOptions:          genericserveroptions.NewEtcdOptions(),
		SecureServingOptions: genericserveroptions.NewSecureServingOptions(),
	}

	cmd := &cobra.Command{
		Use:   "start",
		Short: "run a service-catalog server",
		Run: func(c *cobra.Command, args []string) {
			options.runServer()
		},
	}

	// eventually we pass flags to sub options to configure them.
	// flags := cmd.Flags()

	return cmd
}

// runServer is a method on the options for composition. allows embedding in a higher level options as we do the etcd and serving options.
func (serverOptions ServerOptions) runServer() error {
	fmt.Println("set up the server")
	// options
	if err := serverOptions.SecureServingOptions.MaybeDefaultWithSelfSignedCerts("localhost"); err != nil {
		fmt.Printf("Error creating self-signed certificates: %v", err)
		return err
	}

	// config
	config := genericapiserver.NewConfig()
	secureConfig, err := config.ApplySecureServingOptions(serverOptions.SecureServingOptions)
	if err != nil {
		return err
	}

	completedconfig := secureConfig.Complete()

	// make the server
	server, err := completedconfig.New()
	if err != nil {
		return err
	}

	preparedserver := server.PrepareRun() // post api installation setup? We should have set up the api already?

	stop := make(chan struct{})
	fmt.Println("run the server")
	preparedserver.Run(stop)
	return nil
}
