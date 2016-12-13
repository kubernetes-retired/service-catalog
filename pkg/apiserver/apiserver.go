package apiserver

import (
	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/genericapiserver"
	"k8s.io/kubernetes/pkg/version"
)

// ServiceCatalogAPIServer contains base GenericAPIServer along with
// other configured runtime confiuration
type ServiceCatalogAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

// Config contains our base generic Config along with config specific
// to the service catalog.
type Config struct {
	GenericConfig *genericapiserver.Config
}

// CompletedConfig is an internal type to take advantage of
// typechecking in the type system. mhb does not like it.
type CompletedConfig struct {
	*Config
}

// Complete fills in any fields not set that are required to have
// valid data and can be derived from other fields.
func (c *Config) Complete() CompletedConfig {
	c.GenericConfig.Complete()

	version := version.Get()
	// Setting this var enables the version resource. We should
	// populate the fields of the object from above if we wish to
	// have our own output. Or establish our own version object
	// somewhere else.
	c.GenericConfig.Version = &version

	return CompletedConfig{c}
}

// New creates the server to run.
func (c CompletedConfig) New() (*ServiceCatalogAPIServer, error) {
	// we need to call new on a "completed" config, which we
	// should already have, as this is a 'CompletedConfig' and the
	// only way to get here from there is by Complete()'ing. Thus
	// we skip the complete on the underlying config and go
	// straight to running it's New() method.
	genericServer, err := c.Config.GenericConfig.SkipComplete().New()
	if err != nil {
		return nil, err
	}

	glog.Infoln("Creating the server to run")

	s := &ServiceCatalogAPIServer{
		GenericAPIServer: genericServer,
	}

	return s, nil
}
