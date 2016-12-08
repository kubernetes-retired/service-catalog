package apiserver

import (
	"k8s.io/kubernetes/pkg/genericapiserver"
	//"k8s.io/kubernetes/pkg/version"
)

//
type ServiceCatalogAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

//
type Config struct {
	GenericConfig *genericapiserver.Config
}

// internal type to take advantage of typechecking in the type
// system. I do not understand this in depth. Rolling with it.
type completedConfig struct {
	*Config
}

func (c *Config) Complete() completedConfig {
	c.GenericConfig.Complete()

	// not sure we need this
	// version := version.Get()
	// c.GenericConfig.Version = &version

	return completedConfig{c}
}

func (c completedConfig) New() (*ServiceCatalogAPIServer, error) {
	// we need to call new on a "completed" config, which we
	// should already have, as this is a 'completedConfig' and the
	// only way to get here from there is by Complete()'ing. Thus
	// we skip the complete on the underlying config and go
	// straight to running it's New() method.
	genericServer, err := c.Config.GenericConfig.SkipComplete().New()
	if err != nil {
		return nil, err
	}

	s := &ServiceCatalogAPIServer{
		GenericAPIServer: genericServer,
	}

	return s, nil
}
