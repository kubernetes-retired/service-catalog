// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

// Package v1alpha1 is largely code generated
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/kubernetes-incubator/service-catalog/pkg/apis/settings
// +k8s:defaulter-gen=TypeMeta
// +groupName=settings.servicecatalog.k8s.io
package v1alpha1 // import "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings/v1alpha1"
