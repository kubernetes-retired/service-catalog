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
	// for the http stuff ?
	SecureServingOptions *genericserveroptions.SecureServingOptions

	// for etcd storage ?
	EtcdOptions *genericserveroptions.EtcdOptions
}

const etcdPathPrefix = "/k8s.io/incubator/service-catalog"

func NewCommandServer(out io.Writer) *cobra.Command {
	options := &ServerOptions{
		SecureServingOptions: genericserveroptions.NewSecureServingOptions(),

		EtcdOptions: genericserveroptions.NewEtcdOptions(),
	}

	options.EtcdOptions.StorageConfig.Prefix = etcdPathPrefix

	cmd := &cobra.Command{
		Use:   "start",
		Short: "run a service-catalog server",
		Run: func(c *cobra.Command, args []string) {
			options.runServer()
		},
	}

	// We pass flags object to sub option structs to have them configure
	// themselves. Each options adds it's own command line flags
	// in addition to the flags that are defined above.
	flags := cmd.Flags()
	options.SecureServingOptions.AddFlags(flags)
	options.EtcdOptions.AddFlags(flags)

	return cmd
}

// runServer is a method on the options for composition. allows embedding in a higher level options as we do the etcd and serving options.
func (serverOptions ServerOptions) runServer() error {
	fmt.Println("set up the server")
	// options
	// server configuration options
	fmt.Println("set up serving options")
	if err := serverOptions.SecureServingOptions.MaybeDefaultWithSelfSignedCerts("localhost"); err != nil { // XXX add a flag for the hostname
		fmt.Printf("Error creating self-signed certificates: %v", err)
		return err
	}
	// etcd configuration options
	fmt.Println("set up etcd options")
	if err := serverOptions.EtcdOptions.Validate(); 0 != len(err) {
		return err[0]
	}

	// config
	fmt.Println("set up config object")
	config := genericapiserver.NewConfig()
	secureConfig, err := config.ApplySecureServingOptions(serverOptions.SecureServingOptions)
	if err != nil {
		return err
	}

	completedconfig := secureConfig.Complete()

	// make the server
	fmt.Println("make the server")
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
