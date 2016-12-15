// Package install registers the service-catalog API group
package install

import (
	"k8s.io/kubernetes/pkg/apimachinery/announced"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"
)

func init() {
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:              servicecatalog.GroupName,
			VersionPreferenceOrder: []string{v1alpha1.SchemeGroupVersion.Version},
			ImportPrefix:           "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog",
			// TODO: what does this do?
			RootScopedKinds: nil, // nil is allowed
			// TODO: Do we have 'internal objects'? What is an 'internal object'?
			// mhb: ? broker/catalog/service/instance are our 'internal objects' ?
			AddInternalObjectsToScheme: servicecatalog.AddToScheme, // nil if there are no 'internal objects'
		},
		// TODO what does this do? Is it necessary?
		announced.VersionToSchemeFunc{
			v1alpha1.SchemeGroupVersion.Version: v1alpha1.AddToScheme,
		},
	).Announce().RegisterAndEnable(); err != nil {
		panic(err)
	}
}
