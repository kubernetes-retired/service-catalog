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

// internal type to take advantage of typechecking in the type system
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
	genericServer, err := c.Config.GenericConfig.SkipComplete().New() // completion is done in Complete, no need for a second time
	if err != nil {
		return nil, err
	}

	s := &ServiceCatalogAPIServer{
		GenericAPIServer: genericServer,
	}

	return s, nil
}
