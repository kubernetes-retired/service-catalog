package servicecatalog

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
)

// SDK wrapper around the generated Go client for the Kubernetes Service Catalog
type SDK struct {
	ServiceCatalogClient *clientset.Clientset
}

// ServiceCatalog is the underlying generated Service Catalog versioned interface
// It should be used instead of accessing the client directly.
func (sdk *SDK) ServiceCatalog() v1beta1.ServicecatalogV1beta1Interface {
	return sdk.ServiceCatalogClient.ServicecatalogV1beta1()
}
