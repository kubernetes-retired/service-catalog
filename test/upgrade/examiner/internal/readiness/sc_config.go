package readiness

// ServiceCatalogConfig collects all parameters from env variables required to run upgrade tests
type ServiceCatalogConfig struct {
	ServiceCatalogApiServerName        string `envconfig:"SC_APISERVER"`
	ServiceCatalogControllerServerName string `envconfig:"SC_CONTROLLER"`
	ServiceCatalogNamespace            string `envconfig:"SC_NAMESPACE"`
	TestBrokerName                     string `envconfig:"TB_NAME"`
	TestBrokerNamespace                string `envconfig:"TB_NAMESPACE"`
}
