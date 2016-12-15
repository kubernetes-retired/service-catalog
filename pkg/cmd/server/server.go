package server

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/golang/glog"

	"github.com/kubernetes-incubator/service-catalog/pkg/apiserver"

	//"k8s.io/kubernetes/pkg/api"
	//"k8s.io/kubernetes/pkg/apimachinery/registered"
	"k8s.io/kubernetes/pkg/genericapiserver"
	genericserveroptions "k8s.io/kubernetes/pkg/genericapiserver/options"
)

// ServiceCatalogServerOptions contains the aggregation of configuration structs for
// the service-catalog server. The theory here is that any future user
// of this server will be able to use this options object as a sub
// options of it's own.
type ServiceCatalogServerOptions struct {
	// the runtime configuration of our server
	GenericServerRunOptions *genericserveroptions.ServerRunOptions
	// the https configuration. certs, etc
	SecureServingOptions *genericserveroptions.SecureServingOptions
	// storage with etcd
	EtcdOptions *genericserveroptions.EtcdOptions
	// authn
	AuthenticationOptions *genericserveroptions.DelegatingAuthenticationOptions
	// authz
	AuthorizationOptions *genericserveroptions.DelegatingAuthorizationOptions
}

const (
	// I made this up to match some existing paths. I am not sure if there
	// are any restrictions on the format or structure beyond text
	// separated by slashes.
	etcdPathPrefix = "/k8s.io/incubator/service-catalog"

	// GroupName I made this up. Maybe we'll need it.
	GroupName = "service-catalog.incubator.k8s.io"
)

// NewCommandServer creates a new cobra command to run our server.
func NewCommandServer(out io.Writer) *cobra.Command {
	// initalize our sub options.
	options := &ServiceCatalogServerOptions{
		GenericServerRunOptions: genericserveroptions.NewServerRunOptions(),
		SecureServingOptions:    genericserveroptions.NewSecureServingOptions(),
		EtcdOptions:             genericserveroptions.NewEtcdOptions(),
		AuthenticationOptions:   genericserveroptions.NewDelegatingAuthenticationOptions(),
		AuthorizationOptions:    genericserveroptions.NewDelegatingAuthorizationOptions(),
	}

	// when our etcd resources are created, put them in a location
	// specific to us
	options.EtcdOptions.StorageConfig.Prefix = etcdPathPrefix

	// I have no idea what this line does but I keep seeing it. Do
	// we have our own `GroupName`? Is it like the pathPrefix for
	// etcd?
	//
	// options.EtcdOptions.StorageConfig.Codec = api.Codecs.LegacyCodec(registered.EnabledVersionsForGroup(api.GroupName)...)

	// this is the one thing that this program does. It runs the apiserver.
	cmd := &cobra.Command{
		Short: "run a service-catalog server",
		Run: func(c *cobra.Command, args []string) {
			options.runServer()
		},
	}

	// We pass flags object to sub option structs to have them configure
	// themselves. Each options adds it's own command line flags
	// in addition to the flags that are defined above.
	flags := cmd.Flags()
	// TODO consider an AddFlags() method on our options
	// struct. Will need to import pflag.
	//
	// repeated pattern seems like it should be refactored if all
	// options were of an interface type that specified AddFlags.
	options.GenericServerRunOptions.AddUniversalFlags(flags)
	options.SecureServingOptions.AddFlags(flags)
	options.EtcdOptions.AddFlags(flags)
	options.AuthenticationOptions.AddFlags(flags)
	options.AuthorizationOptions.AddFlags(flags)

	return cmd
}

// runServer is a method on the options for composition. allows embedding in a higher level options as we do the etcd and serving options.
func (serverOptions ServiceCatalogServerOptions) runServer() error {
	glog.Infoln("set up the server")
	// options
	// runtime options
	if err := serverOptions.GenericServerRunOptions.DefaultExternalAddress(serverOptions.SecureServingOptions, nil); err != nil {
		return err
	}
	// server configuration options
	glog.Infoln("set up serving options")
	if err := serverOptions.SecureServingOptions.MaybeDefaultWithSelfSignedCerts(serverOptions.GenericServerRunOptions.AdvertiseAddress.String()); err != nil {
		glog.Errorf("Error creating self-signed certificates: %v", err)
		return err
	}

	// etcd options
	if errs := serverOptions.EtcdOptions.Validate(); len(errs) > 0 {
		glog.Errorln("set up etcd options, do you have `--etcd-servers localhost` set?")
		return errs[0]
	}

	// config
	glog.Infoln("set up config object")
	genericconfig := genericapiserver.NewConfig().ApplyOptions(serverOptions.GenericServerRunOptions)
	// these are all mutators of each specific suboption in serverOptions object.
	// this repeated pattern seems like we could refactor
	if _, err := genericconfig.ApplySecureServingOptions(serverOptions.SecureServingOptions); err != nil {
		glog.Errorln(err)
		return err
	}

	// need to figure out what's throwing the `missing clientCA file` err
	/*
		if _, err := genericconfig.ApplyDelegatingAuthenticationOptions(serverOptions.AuthenticationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/
	// having this enabled causes the server to crash for any call
	/*
		if _, err := genericconfig.ApplyDelegatingAuthorizationOptions(serverOptions.AuthorizationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/

	// configure our own apiserver using the preconfigured genericApiServer
	config := apiserver.Config{
		GenericConfig: genericconfig,
	}

	// finish config
	completedconfig := config.Complete()

	// make the server
	glog.Infoln("make the server")
	server, err := completedconfig.New()
	if err != nil {
		return err
	}

	// I don't like this. We're reaching in too far to call things.
	preparedserver := server.GenericAPIServer.PrepareRun() // post api installation setup? We should have set up the api already?

	stop := make(chan struct{})
	glog.Infoln("run the server")
	preparedserver.Run(stop)
	return nil
}

/*
type restOptionsFactory struct {
	storageConfig *storagebackend.Config
}
*/
/*
func (f restOptionsFactory) NewFor(resource schema.GroupResource) generic.RESTOptions {
	return generic.RESTOptions{
		StorageConfig:           f.storageConfig,
		Decorator:               registry.StorageWithCacher,
		DeleteCollectionWorkers: 1,
		EnableGarbageCollection: false,
		ResourcePrefix:          f.storageConfig.Prefix + "/" + resource.Group + "/" + resource.Resource,
	}
}*/
