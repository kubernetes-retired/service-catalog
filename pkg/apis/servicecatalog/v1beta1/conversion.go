/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/filter"
	"k8s.io/apimachinery/pkg/labels"
)

// These functions are used for field selectors. They are only needed if
// field selection is made available for types, hence we only have them for
// ServicePlan and ServiceClass. While they are identical, it's clearer to
// use different functions from the get go.

// ClusterServicePlanFieldLabelConversionFunc does not convert anything, just returns
// what it's given for the supported fields, and errors for unsupported.
func ClusterServicePlanFieldLabelConversionFunc(label, value string) (string, string, error) {
	switch label {
	case "spec.externalID",
		"spec.externalName",
		"spec.clusterServiceBrokerName",
		"spec.clusterServiceClassRef.name":
		return label, value, nil
	default:
		return "", "", fmt.Errorf("field label not supported: %s", label)
	}
}

// ClusterServiceClassFieldLabelConversionFunc does not convert anything, just returns
// what it's given for the supported fields, and errors for unsupported.
func ClusterServiceClassFieldLabelConversionFunc(label, value string) (string, string, error) {
	switch label {
	case "spec.externalID",
		"spec.externalName",
		"spec.clusterServiceBrokerName":
		return label, value, nil
	default:
		return "", "", fmt.Errorf("field label not supported: %s", label)
	}
}

// ServiceInstanceFieldLabelConversionFunc does not convert anything, just returns
// what it's given for the supported fields, and errors for unsupported.
func ServiceInstanceFieldLabelConversionFunc(label, value string) (string, string, error) {
	switch label {
	case "spec.clusterServiceClassRef.name",
		"spec.clusterServicePlanRef.name":
		return label, value, nil
	default:
		return "", "", fmt.Errorf("field label not supported: %s", label)
	}
}

// ConvertClusterServiceClassToProperties takes a Service Class and pulls out the
// properties we support for filtering, converting them into a map in the
// expected format.
func ConvertClusterServiceClassToProperties(serviceClass *ClusterServiceClass) filter.Properties {
	if serviceClass == nil {
		return labels.Set{}
	}
	return labels.Set{
		FilterName:             serviceClass.Name,
		FilterSpecExternalName: serviceClass.Spec.ExternalName,
		FilterSpecExternalID:   serviceClass.Spec.ExternalID,
	}
}

// ConvertServicePlanToProperties takes a Service Plan and pulls out the
// properties we support for filtering, converting them into a map in the
// expected format.
func ConvertClusterServicePlanToProperties(servicePlan *ClusterServicePlan) filter.Properties {
	if servicePlan == nil {
		return labels.Set{}
	}
	return labels.Set{
		FilterName:                        servicePlan.Name,
		FilterSpecExternalName:            servicePlan.Spec.ExternalName,
		FilterSpecExternalID:              servicePlan.Spec.ExternalID,
		FilterSpecClusterServiceClassName: servicePlan.Spec.ClusterServiceClassRef.Name,
	}
}
