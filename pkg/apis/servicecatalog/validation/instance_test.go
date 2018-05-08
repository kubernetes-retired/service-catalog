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

package validation

import (
	"reflect"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

const (
	clusterServiceClassExternalName = "test-clusterserviceclass"
	clusterServiceClassExternalID   = "test-clusterserviceclass-ext-id"
	clusterServicePlanExternalName  = "test-clusterserviceplan"
	clusterServicePlanExternalID    = "test-clusterserviceplan-ext-id"
	clusterServiceClassName         = "test-k8s-serviceclass"
	clusterServicePlanName          = "test-k8s-plan-name"

	serviceClassExternalName = "test-serviceclass"
	serviceClassExternalID   = "test-serviceclass-ext-id"
	servicePlanExternalName  = "test-serviceplan"
	servicePlanExternalID    = "test-serviceplan-ext-id"
	serviceClassName         = "test-k8s-serviceclass"
	servicePlanName          = "test-k8s-plan-name"
)

func validPlanReferenceClusterServiceExternalName() servicecatalog.PlanReference {
	return servicecatalog.PlanReference{
		ClusterServiceClassExternalName: clusterServiceClassExternalName,
		ClusterServicePlanExternalName:  clusterServicePlanExternalName,
	}
}

func validPlanReferenceClusterServiceExternalID() servicecatalog.PlanReference {
	return servicecatalog.PlanReference{
		ClusterServiceClassExternalID: clusterServiceClassExternalID,
		ClusterServicePlanExternalID:  clusterServicePlanExternalID,
	}
}

func validPlanReferenceClusterK8S() servicecatalog.PlanReference {
	return servicecatalog.PlanReference{
		ClusterServiceClassName: clusterServiceClassName,
		ClusterServicePlanName:  clusterServicePlanName,
	}
}

func validPlanReferenceServiceExternalName() servicecatalog.PlanReference {
	return servicecatalog.PlanReference{
		ServiceClassExternalName: serviceClassExternalName,
		ServicePlanExternalName:  servicePlanExternalName,
	}
}

func validPlanReferenceServiceExternalID() servicecatalog.PlanReference {
	return servicecatalog.PlanReference{
		ServiceClassExternalID: serviceClassExternalID,
		ServicePlanExternalID:  servicePlanExternalID,
	}
}

func validPlanReferenceK8S() servicecatalog.PlanReference {
	return servicecatalog.PlanReference{
		ServiceClassName: serviceClassName,
		ServicePlanName:  servicePlanName,
	}
}

func validServiceInstanceForCreateClusterPlanRef() *servicecatalog.ServiceInstance {
	return &servicecatalog.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-instance",
			Namespace:  "test-ns",
			Generation: 1,
		},
		Spec: servicecatalog.ServiceInstanceSpec{
			PlanReference: validPlanReferenceClusterServiceExternalName(),
		},
		Status: servicecatalog.ServiceInstanceStatus{
			DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusNotRequired,
		},
	}
}

func validServiceInstanceForCreateNamespacedPlanRef() *servicecatalog.ServiceInstance {
	return &servicecatalog.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-instance",
			Namespace:  "test-ns",
			Generation: 1,
		},
		Spec: servicecatalog.ServiceInstanceSpec{
			PlanReference: validPlanReferenceServiceExternalName(),
		},
		Status: servicecatalog.ServiceInstanceStatus{
			DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusNotRequired,
		},
	}
}

func validClusterRefServiceInstance() *servicecatalog.ServiceInstance {
	instance := validServiceInstanceForCreateClusterPlanRef()
	instance.Spec.ClusterServiceClassRef = &servicecatalog.ClusterObjectReference{}
	instance.Spec.ClusterServicePlanRef = &servicecatalog.ClusterObjectReference{}
	return instance
}

func validNamespacedRefServiceInstance() *servicecatalog.ServiceInstance {
	instance := validServiceInstanceForCreateNamespacedPlanRef()
	instance.Spec.ServiceClassRef = &servicecatalog.LocalObjectReference{}
	instance.Spec.ServicePlanRef = &servicecatalog.LocalObjectReference{}
	return instance
}

func validServiceInstanceWithInProgressProvision() *servicecatalog.ServiceInstance {
	instance := validClusterRefServiceInstance()
	instance.Generation = 2
	instance.Status.ReconciledGeneration = 1
	instance.Status.CurrentOperation = servicecatalog.ServiceInstanceOperationProvision
	now := metav1.Now()
	instance.Status.OperationStartTime = &now
	instance.Status.InProgressProperties = validServiceInstancePropertiesStateClusterPlan()
	return instance
}

func validServiceInstanceWithInProgressDeprovision() *servicecatalog.ServiceInstance {
	instance := validClusterRefServiceInstance()
	instance.Generation = 2
	instance.Status.ReconciledGeneration = 1
	instance.Status.CurrentOperation = servicecatalog.ServiceInstanceOperationDeprovision
	now := metav1.Now()
	instance.Status.OperationStartTime = &now
	instance.Status.InProgressProperties = validServiceInstancePropertiesStateClusterPlan()
	instance.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
	return instance
}

func validServiceInstancePropertiesStateClusterPlan() *servicecatalog.ServiceInstancePropertiesState {
	return &servicecatalog.ServiceInstancePropertiesState{
		ClusterServicePlanExternalName: "plan-name",
		ClusterServicePlanExternalID:   "plan-id",
		Parameters:                     &runtime.RawExtension{Raw: []byte("a: 1\nb: \"2\"")},
		ParametersChecksum:             "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}
}

func validServiceInstancePropertiesStateNamespacedPlan() *servicecatalog.ServiceInstancePropertiesState {
	return &servicecatalog.ServiceInstancePropertiesState{
		ServicePlanExternalName: "ns-plan-name",
		ServicePlanExternalID:   "ns-plan-id",
		Parameters:              &runtime.RawExtension{Raw: []byte("a: 1\nb: \"2\"")},
		ParametersChecksum:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}
}

func TestValidateServiceInstance(t *testing.T) {
	cases := []struct {
		name     string
		instance *servicecatalog.ServiceInstance
		create   bool
		valid    bool
	}{
		{
			name:     "valid cluster ref service instance",
			instance: validClusterRefServiceInstance(),
			valid:    true,
		},
		{
			name:     "valid ns ref service instance",
			instance: validClusterRefServiceInstance(),
			valid:    true,
		},
		{
			name: "invalid -- cluster & ns ref",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ServiceClassRef = &servicecatalog.LocalObjectReference{}
				i.Spec.ServicePlanRef = &servicecatalog.LocalObjectReference{}
				return i
			}(),
			valid: false,
		},
		{
			name: "valid planName",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServicePlanExternalName = "9651.JVHbebe"
				return i
			}(),
			valid: true,
		},
		{
			name: "missing namespace",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Namespace = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "missing clusterServiceClassExternalName and clusterServiceClassName",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServiceClassExternalName = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "invalid clusterServiceClassExternalName",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServiceClassExternalName = "oing20&)*^&"
				return i
			}(),
			valid: false,
		},
		{
			name: "missing clusterServicePlanExternalName",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServicePlanExternalName = ""
				return i
			}(),
			valid: true, // plan may be picked by defaultserviceplan admission controller
		},
		{
			name: "invalid clusterServicePlanExternalName",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServicePlanExternalName = "9651_JVHbebe"
				return i
			}(),
			valid: false,
		},
		{
			name: "valid parametersFrom",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ParametersFrom =
					[]servicecatalog.ParametersFromSource{
						{SecretKeyRef: &servicecatalog.SecretKeyReference{Name: "test-key-name", Key: "test-key"}}}
				return i
			}(),
			valid: true,
		},
		{
			name: "missing key reference in parametersFrom",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ParametersFrom =
					[]servicecatalog.ParametersFromSource{{SecretKeyRef: nil}}
				return i
			}(),
			valid: false,
		},
		{
			name: "key name is missing in parametersFrom",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ParametersFrom =
					[]servicecatalog.ParametersFromSource{
						{SecretKeyRef: &servicecatalog.SecretKeyReference{Name: "", Key: "test-key"}}}
				return i
			}(),
			valid: false,
		},
		{
			name: "key is missing in parametersFrom",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ParametersFrom =
					[]servicecatalog.ParametersFromSource{
						{SecretKeyRef: &servicecatalog.SecretKeyReference{Name: "test-key-name", Key: ""}}}
				return i
			}(),
			valid: false,
		},
		{
			name:     "valid with in-progress provision",
			instance: validServiceInstanceWithInProgressProvision(),
			valid:    true,
		},
		{
			name: "valid with in-progress update",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.CurrentOperation = servicecatalog.ServiceInstanceOperationUpdate
				return i
			}(),
			valid: true,
		},
		{
			name: "valid with in-progress deprovision",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.CurrentOperation = servicecatalog.ServiceInstanceOperationDeprovision
				return i
			}(),
			valid: true,
		},
		{
			name: "invalid current operation",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.CurrentOperation = servicecatalog.ServiceInstanceOperation("bad-operation")
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress without updated generation",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.ReconciledGeneration = i.Generation
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress with missing OperationStartTime",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.OperationStartTime = nil
				return i
			}(),
			valid: false,
		},
		{
			name: "not in-progress with present OperationStartTime",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				now := metav1.Now()
				i.Status.OperationStartTime = &now
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress with condition ready/true",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.Conditions = []servicecatalog.ServiceInstanceCondition{
					{
						Type:   servicecatalog.ServiceInstanceConditionReady,
						Status: servicecatalog.ConditionTrue,
					},
				}
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress with condition ready/false",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.Conditions = []servicecatalog.ServiceInstanceCondition{
					{
						Type:   servicecatalog.ServiceInstanceConditionReady,
						Status: servicecatalog.ConditionFalse,
					},
				}
				return i
			}(),
			valid: true,
		},
		{
			name: "in-progress provision with missing InProgressProperties",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties = nil
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress update with missing InProgressProperties",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.CurrentOperation = servicecatalog.ServiceInstanceOperationUpdate
				i.Status.InProgressProperties = nil
				return i
			}(),
			valid: false,
		},
		{
			name: "not in-progress with present InProgressProperties",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.InProgressProperties = validServiceInstancePropertiesStateClusterPlan()
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress properties with no external plan name",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.ClusterServicePlanExternalName = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress properties with no external plan ID",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.ClusterServicePlanExternalID = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "valid in-progress properties with no parameters",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.Parameters = nil
				i.Status.InProgressProperties.ParametersChecksum = ""
				return i
			}(),
			valid: true,
		},
		{
			name: "in-progress properties parameters with missing parameters checksum",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.ParametersChecksum = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress properties parameters checksum with missing parameters",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.Parameters = nil
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress properties parameters with missing raw",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.Parameters.Raw = []byte{}
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress properties parameters with malformed yaml",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.Parameters.Raw = []byte("bad yaml")
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress properties parameters checksum too small",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.ParametersChecksum = "0123456"
				return i
			}(),
			valid: false,
		},
		{
			name: "in-progress properties parameters checksum malformed",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.InProgressProperties.ParametersChecksum = "not hex"
				return i
			}(),
			valid: false,
		},
		{
			name: "valid external properties cluster plan",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				return i
			}(),
			valid: true,
		},
		{
			name: "valid external properties namespaced plan",
			instance: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateNamespacedPlan()
				return i
			}(),
			valid: true,
		},
		{
			name: "invalid external properties cluster + ns plan",
			instance: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				props := validServiceInstancePropertiesStateClusterPlan()
				props.ServicePlanExternalName = "ns-plan-name"
				props.ServicePlanExternalID = "ns-plan-id"
				i.Status.ExternalProperties = props
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties with no external cluster plan name",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.ClusterServicePlanExternalName = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties with no external cluster plan ID",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.ClusterServicePlanExternalID = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties with no external namespaced plan name",
			instance: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateNamespacedPlan()
				i.Status.ExternalProperties.ServicePlanExternalName = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties with no external namespaced plan ID",
			instance: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateNamespacedPlan()
				i.Status.ExternalProperties.ServicePlanExternalID = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "valid external properties with no parameters",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.Parameters = nil
				i.Status.ExternalProperties.ParametersChecksum = ""
				return i
			}(),
			valid: true,
		},
		{
			name: "external properties parameters with missing parameters checksum",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.ParametersChecksum = ""
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties parameters checksum with missing parameters",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.Parameters = nil
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties parameters with missing raw",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.Parameters.Raw = []byte{}
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties parameters with malformed yaml",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.Parameters.Raw = []byte("bad yaml")
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties parameters checksum too small",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.ParametersChecksum = "0123456"
				return i
			}(),
			valid: false,
		},
		{
			name: "external properties parameters checksum malformed",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ExternalProperties = validServiceInstancePropertiesStateClusterPlan()
				i.Status.ExternalProperties.ParametersChecksum = "not hex"
				return i
			}(),
			valid: false,
		},
		{
			name:     "valid create with cluster plan ref",
			instance: validServiceInstanceForCreateClusterPlanRef(),
			create:   true,
			valid:    true,
		},
		{
			name:     "valid create with namespaced plan ref",
			instance: validServiceInstanceForCreateClusterPlanRef(),
			create:   true,
			valid:    true,
		},
		{
			name: "valid create with k8s name -- cluster ref",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceForCreateClusterPlanRef()
				i.Spec.ClusterServiceClassExternalName = ""
				i.Spec.ClusterServicePlanExternalName = ""
				i.Spec.ClusterServiceClassName = clusterServiceClassName
				i.Spec.ClusterServicePlanName = clusterServicePlanName
				return i
			}(),
			create: true,
			valid:  true,
		},
		{
			name: "valid create with k8s name -- namespaced ref",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceForCreateNamespacedPlanRef()
				i.Spec.ServiceClassExternalName = ""
				i.Spec.ServicePlanExternalName = ""
				i.Spec.ServiceClassName = serviceClassName
				i.Spec.ServicePlanName = servicePlanName
				return i
			}(),
			create: true,
			valid:  true,
		},
		{
			name: "create with operation in-progress",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceForCreateClusterPlanRef()
				i.Status.CurrentOperation = servicecatalog.ServiceInstanceOperationProvision
				return i
			}(),
			create: true,
			valid:  false,
		},
		{
			name: "create with invalid reconciled generation",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceForCreateClusterPlanRef()
				i.Status.ReconciledGeneration = 1
				return i
			}(),
			create: true,
			valid:  false,
		},
		{
			name: "update with invalid reconciled generation",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.ReconciledGeneration = 2
				return i
			}(),
			create: false,
			valid:  false,
		},
		{
			name: "in-progress operation with missing cluster & namespaced  service class ref",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Spec.ClusterServiceClassRef = nil
				i.Spec.ServiceClassRef = nil
				return i
			}(),
			create: false,
			valid:  false,
		},
		{
			name: "in-progress provision with missing cluster & namespaced service plan ref",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Spec.ClusterServicePlanRef = nil
				i.Spec.ServicePlanRef = nil
				return i
			}(),
			create: false,
			valid:  false,
		},
		{
			name:     "valid in-progress deprovision",
			instance: validServiceInstanceWithInProgressDeprovision(),
			create:   false,
			valid:    true,
		},
		{
			name: "in-progress deprovision with missing cluster & namespaced service plan ref",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressDeprovision()
				i.Spec.ClusterServicePlanRef = nil
				i.Spec.ServicePlanRef = nil
				return i
			}(),
			create: false,
			valid:  true,
		},
		{
			name: "in-progress deprovision with missing external properties",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressDeprovision()
				i.Status.ExternalProperties = nil
				return i
			}(),
			create: false,
			valid:  true,
		},
		{
			name: "in-progress deprovision with missing external properties plan ID",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressDeprovision()
				i.Status.ExternalProperties.ClusterServicePlanExternalID = ""
				i.Status.ExternalProperties.ServicePlanExternalID = ""
				return i
			}(),
			create: false,
			// not valid because ClusterServicePlanExternalID is required when ExternalProperties is present
			valid: false,
		},
		{
			name: "in-progress deprovision with missing cluster & namespaced service plan ref and external properties",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressDeprovision()
				i.Spec.ClusterServicePlanRef = nil
				i.Spec.ServicePlanRef = nil
				i.Status.ExternalProperties = nil
				return i
			}(),
			create: false,
			valid:  false,
		},
		{
			name: "in-progress deprovision with missing cluster & namespaced service plan ref and external properties cluster & namespaced plan ID",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressDeprovision()
				i.Spec.ClusterServicePlanRef = nil
				i.Status.ExternalProperties.ClusterServicePlanExternalID = ""
				i.Status.ExternalProperties.ServicePlanExternalID = ""
				return i
			}(),
			create: false,
			valid:  false,
		},
		{
			name: "external and k8s name specified in cluster Spec.PlanReference",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServiceClassName = "can not have this here"
				return i
			}(),
			create: true,
			valid:  false,
		},
		{
			name: "external and k8s name specified in namespaced Spec.PlanReference",
			instance: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Spec.ServiceClassName = "can not have this here"
				return i
			}(),
			create: true,
			valid:  false,
		},
		{
			name: "failed provision starting orphan mitigation",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.OperationStartTime = nil
				i.Status.OrphanMitigationInProgress = true
				i.Status.Conditions = []servicecatalog.ServiceInstanceCondition{
					{
						Type:   servicecatalog.ServiceInstanceConditionReady,
						Status: servicecatalog.ConditionFalse,
					},
					{
						Type:   servicecatalog.ServiceInstanceConditionFailed,
						Status: servicecatalog.ConditionTrue,
					},
				}
				return i
			}(),
			valid: true,
		},
		{
			name: "in-progress orphan mitigation",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceWithInProgressProvision()
				i.Status.OrphanMitigationInProgress = true
				i.Status.Conditions = []servicecatalog.ServiceInstanceCondition{
					{
						Type:   servicecatalog.ServiceInstanceConditionReady,
						Status: servicecatalog.ConditionFalse,
					},
					{
						Type:   servicecatalog.ServiceInstanceConditionFailed,
						Status: servicecatalog.ConditionTrue,
					},
				}
				return i
			}(),
			valid: true,
		},
		{
			name: "required deprovision status on create",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceForCreateClusterPlanRef()
				i.Status.DeprovisionStatus = servicecatalog.ServiceInstanceDeprovisionStatusRequired
				return i
			}(),
			create: true,
			valid:  false,
		},
		{
			name: "succeeded deprovision status on create",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceForCreateClusterPlanRef()
				i.Status.DeprovisionStatus = servicecatalog.ServiceInstanceDeprovisionStatusSucceeded
				return i
			}(),
			create: true,
			valid:  false,
		},
		{
			name: "failed deprovision status on create",
			instance: func() *servicecatalog.ServiceInstance {
				i := validServiceInstanceForCreateClusterPlanRef()
				i.Status.DeprovisionStatus = servicecatalog.ServiceInstanceDeprovisionStatusFailed
				return i
			}(),
			create: true,
			valid:  false,
		},
		{
			name: "invalid deprovision status on update",
			instance: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Status.DeprovisionStatus = servicecatalog.ServiceInstanceDeprovisionStatus("bad-deprovision-status")
				return i
			}(),
			valid: false,
		},
	}

	for _, tc := range cases {
		errs := internalValidateServiceInstance(tc.instance, tc.create)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}

func TestInternalValidateServiceInstanceUpdateAllowed(t *testing.T) {
	cases := []struct {
		name             string
		specChange       bool
		onGoingOperation bool
		valid            bool
	}{
		{
			name:             "spec change when no on-going operation",
			specChange:       true,
			onGoingOperation: false,
			valid:            true,
		},
		{
			name:             "spec change when on-going operation",
			specChange:       true,
			onGoingOperation: true,
			valid:            false,
		},
		{
			name:             "meta change when no on-going operation",
			specChange:       false,
			onGoingOperation: false,
			valid:            true,
		},
		{
			name:             "meta change when on-going operation",
			specChange:       false,
			onGoingOperation: true,
			valid:            true,
		},
	}

	for _, tc := range cases {
		oldInstance := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: "test-ns",
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference: servicecatalog.PlanReference{
					ClusterServiceClassExternalName: clusterServiceClassExternalName,
					ClusterServicePlanExternalName:  clusterServicePlanExternalName,
				},
			},
		}
		oldInstance.Generation = 1
		if tc.onGoingOperation {
			oldInstance.Status.CurrentOperation = servicecatalog.ServiceInstanceOperationProvision
		}

		newInstance := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: "test-ns",
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference: servicecatalog.PlanReference{
					ClusterServiceClassExternalName: clusterServiceClassExternalName,
					ClusterServicePlanExternalName:  clusterServicePlanExternalName,
				},
			},
		}
		if tc.specChange {
			newInstance.Generation = oldInstance.Generation + 1
		} else {
			newInstance.Generation = oldInstance.Generation
		}

		errs := internalValidateServiceInstanceUpdateAllowed(newInstance, oldInstance)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}

	for _, tc := range cases {
		oldInstance := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: "test-ns",
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference: servicecatalog.PlanReference{
					ServiceClassExternalName: serviceClassExternalName,
					ServicePlanExternalName:  servicePlanExternalName,
				},
			},
		}
		oldInstance.Generation = 1
		if tc.onGoingOperation {
			oldInstance.Status.CurrentOperation = servicecatalog.ServiceInstanceOperationProvision
		}

		newInstance := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: "test-ns",
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference: servicecatalog.PlanReference{
					ServiceClassExternalName: serviceClassExternalName,
					ServicePlanExternalName:  servicePlanExternalName,
				},
			},
		}
		if tc.specChange {
			newInstance.Generation = oldInstance.Generation + 1
		} else {
			newInstance.Generation = oldInstance.Generation
		}

		errs := internalValidateServiceInstanceUpdateAllowed(newInstance, oldInstance)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}

func TestInternalValidateServiceInstanceUpdateAllowedForClusterPlanChange(t *testing.T) {
	newClusterPlanExternalName := servicecatalog.PlanReference{
		ClusterServiceClassExternalName: clusterServiceClassExternalName,
		ClusterServicePlanExternalName:  "new-plan",
	}
	newClusterPlanExternalID := servicecatalog.PlanReference{
		ClusterServiceClassExternalName: clusterServiceClassExternalID,
		ClusterServicePlanExternalID:    "new-plan",
	}
	newClusterPlanK8sName := servicecatalog.PlanReference{
		ClusterServiceClassName: clusterServiceClassName,
		ClusterServicePlanName:  "new-plan",
	}

	cases := []struct {
		name       string
		oldPlan    servicecatalog.PlanReference
		newPlan    servicecatalog.PlanReference
		newPlanRef *servicecatalog.ClusterObjectReference
		valid      bool
	}{
		{
			name:       "valid cluster plan change via external name",
			oldPlan:    validPlanReferenceClusterServiceExternalName(),
			newPlan:    newClusterPlanExternalName,
			newPlanRef: nil,
			valid:      true,
		},
		{
			name:       "valid cluster plan change via external id",
			oldPlan:    validPlanReferenceClusterServiceExternalID(),
			newPlan:    newClusterPlanExternalID,
			newPlanRef: nil,
			valid:      true,
		},
		{
			name:       "valid cluster plan change via k8s name",
			oldPlan:    validPlanReferenceClusterK8S(),
			newPlan:    newClusterPlanK8sName,
			newPlanRef: nil,
			valid:      true,
		},
		{
			name:       "cluster plan ref not cleared for change via external name",
			oldPlan:    validPlanReferenceClusterServiceExternalName(),
			newPlan:    newClusterPlanExternalName,
			newPlanRef: &servicecatalog.ClusterObjectReference{},
			valid:      false,
		},
		{
			name:       "cluster plan ref not cleared for change via external id",
			oldPlan:    validPlanReferenceClusterServiceExternalID(),
			newPlan:    newClusterPlanExternalID,
			newPlanRef: &servicecatalog.ClusterObjectReference{},
			valid:      false,
		},
		{
			name:       "cluster plan ref not cleared for change via k8s name",
			oldPlan:    validPlanReferenceClusterK8S(),
			newPlan:    newClusterPlanK8sName,
			newPlanRef: &servicecatalog.ClusterObjectReference{},
			valid:      false,
		},
		{
			name:       "no cluster plan change",
			oldPlan:    validPlanReferenceClusterServiceExternalName(),
			newPlan:    validPlanReferenceClusterServiceExternalName(),
			newPlanRef: &servicecatalog.ClusterObjectReference{},
			valid:      true,
		},
	}

	for _, tc := range cases {
		oldInstance := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: "test-ns",
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference:          tc.oldPlan,
				ClusterServiceClassRef: &servicecatalog.ClusterObjectReference{},
				ClusterServicePlanRef:  &servicecatalog.ClusterObjectReference{},
			},
		}

		newInstance := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: "test-ns",
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference:          tc.newPlan,
				ClusterServiceClassRef: &servicecatalog.ClusterObjectReference{},
				ClusterServicePlanRef:  tc.newPlanRef,
			},
		}

		errs := internalValidateServiceInstanceUpdateAllowed(newInstance, oldInstance)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}

func TestInternalValidateServiceInstanceUpdateAllowedForPlanChange(t *testing.T) {
	newPlanExternalName := servicecatalog.PlanReference{
		ServiceClassExternalName: serviceClassExternalName,
		ServicePlanExternalName:  "new-plan",
	}
	newPlanExternalID := servicecatalog.PlanReference{
		ServiceClassExternalName: serviceClassExternalID,
		ServicePlanExternalID:    "new-plan",
	}
	newPlanK8sName := servicecatalog.PlanReference{
		ServiceClassName: serviceClassName,
		ServicePlanName:  "new-plan",
	}

	cases := []struct {
		name       string
		oldPlan    servicecatalog.PlanReference
		newPlan    servicecatalog.PlanReference
		newPlanRef *servicecatalog.LocalObjectReference
		valid      bool
	}{
		{
			name:       "valid namespaced plan change via external name",
			oldPlan:    validPlanReferenceServiceExternalName(),
			newPlan:    newPlanExternalName,
			newPlanRef: nil,
			valid:      true,
		},
		{
			name:       "valid namespaced plan change via external id",
			oldPlan:    validPlanReferenceServiceExternalID(),
			newPlan:    newPlanExternalID,
			newPlanRef: nil,
			valid:      true,
		},
		{
			name:       "valid namespaced plan change via k8s name",
			oldPlan:    validPlanReferenceK8S(),
			newPlan:    newPlanK8sName,
			newPlanRef: nil,
			valid:      true,
		},
		{
			name:       "namespaced plan ref not cleared for change via external name",
			oldPlan:    validPlanReferenceServiceExternalName(),
			newPlan:    newPlanExternalName,
			newPlanRef: &servicecatalog.LocalObjectReference{},
			valid:      false,
		},
		{
			name:       "plan ref not cleared for change via external id",
			oldPlan:    validPlanReferenceServiceExternalID(),
			newPlan:    newPlanExternalID,
			newPlanRef: &servicecatalog.LocalObjectReference{},
			valid:      false,
		},
		{
			name:       "plan ref not cleared for change via k8s name",
			oldPlan:    validPlanReferenceK8S(),
			newPlan:    newPlanK8sName,
			newPlanRef: &servicecatalog.LocalObjectReference{},
			valid:      false,
		},
		{
			name:       "no plan change",
			oldPlan:    validPlanReferenceServiceExternalName(),
			newPlan:    validPlanReferenceServiceExternalName(),
			newPlanRef: &servicecatalog.LocalObjectReference{},
			valid:      true,
		},
	}

	for _, tc := range cases {
		oldInstance := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: "test-ns",
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference:   tc.oldPlan,
				ServiceClassRef: &servicecatalog.LocalObjectReference{},
				ServicePlanRef:  &servicecatalog.LocalObjectReference{},
			},
		}

		newInstance := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-instance",
				Namespace: "test-ns",
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference:   tc.newPlan,
				ServiceClassRef: &servicecatalog.LocalObjectReference{},
				ServicePlanRef:  tc.newPlanRef,
			},
		}

		errs := internalValidateServiceInstanceUpdateAllowed(newInstance, oldInstance)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}

func TestValidateServiceInstanceStatusUpdate(t *testing.T) {
	now := metav1.Now()
	cases := []struct {
		name  string
		old   *servicecatalog.ServiceInstanceStatus
		new   *servicecatalog.ServiceInstanceStatus
		valid bool
		err   string // Error string to match against if error expected
	}{
		{
			name: "Start async op",
			old: &servicecatalog.ServiceInstanceStatus{
				AsyncOpInProgress: false,
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			new: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:     servicecatalog.ServiceInstanceOperationProvision,
				OperationStartTime:   &now,
				InProgressProperties: validServiceInstancePropertiesStateClusterPlan(),
				AsyncOpInProgress:    true,
				DeprovisionStatus:    servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			valid: true,
			err:   "",
		},
		{
			name: "Complete async op",
			old: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:     servicecatalog.ServiceInstanceOperationProvision,
				OperationStartTime:   &now,
				InProgressProperties: validServiceInstancePropertiesStateClusterPlan(),
				AsyncOpInProgress:    true,
				DeprovisionStatus:    servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			new: &servicecatalog.ServiceInstanceStatus{
				AsyncOpInProgress: false,
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			valid: true,
			err:   "",
		},
		{
			name: "ServiceInstanceConditionReady can not be true if operation is ongoing",
			old: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation: "",
				Conditions: []servicecatalog.ServiceInstanceCondition{{
					Type:   servicecatalog.ServiceInstanceConditionReady,
					Status: servicecatalog.ConditionFalse,
				}},
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			new: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:     servicecatalog.ServiceInstanceOperationProvision,
				OperationStartTime:   &now,
				InProgressProperties: validServiceInstancePropertiesStateClusterPlan(),
				Conditions: []servicecatalog.ServiceInstanceCondition{{
					Type:   servicecatalog.ServiceInstanceConditionReady,
					Status: servicecatalog.ConditionTrue,
				}},
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			valid: false,
			err:   "operation in progress",
		},
		{
			name: "ServiceInstanceConditionReady can be true if operation is completed",
			old: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:     servicecatalog.ServiceInstanceOperationProvision,
				OperationStartTime:   &now,
				InProgressProperties: validServiceInstancePropertiesStateClusterPlan(),
				Conditions: []servicecatalog.ServiceInstanceCondition{{
					Type:   servicecatalog.ServiceInstanceConditionReady,
					Status: servicecatalog.ConditionFalse,
				}},
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			new: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation: "",
				Conditions: []servicecatalog.ServiceInstanceCondition{{
					Type:   servicecatalog.ServiceInstanceConditionReady,
					Status: servicecatalog.ConditionTrue,
				}},
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			valid: true,
			err:   "",
		},
		{
			name: "Update non-ready instance condition during operation",
			old: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:     servicecatalog.ServiceInstanceOperationProvision,
				OperationStartTime:   &now,
				InProgressProperties: validServiceInstancePropertiesStateClusterPlan(),
				Conditions:           []servicecatalog.ServiceInstanceCondition{{Status: servicecatalog.ConditionFalse}},
				DeprovisionStatus:    servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			new: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:     servicecatalog.ServiceInstanceOperationProvision,
				OperationStartTime:   &now,
				InProgressProperties: validServiceInstancePropertiesStateClusterPlan(),
				Conditions:           []servicecatalog.ServiceInstanceCondition{{Status: servicecatalog.ConditionTrue}},
				DeprovisionStatus:    servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			valid: true,
			err:   "",
		},
		{
			name: "Update non-ready instance condition outside of operation",
			old: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:  "",
				Conditions:        []servicecatalog.ServiceInstanceCondition{{Status: servicecatalog.ConditionFalse}},
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			new: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:  "",
				Conditions:        []servicecatalog.ServiceInstanceCondition{{Status: servicecatalog.ConditionTrue}},
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			valid: true,
			err:   "",
		},
		{
			name: "Update instance condition to ready status and finish operation",
			old: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:     servicecatalog.ServiceInstanceOperationProvision,
				OperationStartTime:   &now,
				InProgressProperties: &servicecatalog.ServiceInstancePropertiesState{},
				Conditions:           []servicecatalog.ServiceInstanceCondition{{Status: servicecatalog.ConditionFalse}},
				DeprovisionStatus:    servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			new: &servicecatalog.ServiceInstanceStatus{
				CurrentOperation:  "",
				Conditions:        []servicecatalog.ServiceInstanceCondition{{Status: servicecatalog.ConditionTrue}},
				DeprovisionStatus: servicecatalog.ServiceInstanceDeprovisionStatusRequired,
			},
			valid: true,
			err:   "",
		},
	}

	for _, tc := range cases {
		old := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-instance",
				Namespace:  "test-ns",
				Generation: 2,
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference: servicecatalog.PlanReference{
					ClusterServiceClassExternalName: clusterServiceClassExternalName,
					ClusterServicePlanExternalName:  clusterServicePlanExternalName,
				},
				ClusterServiceClassRef: &servicecatalog.ClusterObjectReference{},
				ClusterServicePlanRef:  &servicecatalog.ClusterObjectReference{},
			},
			Status: *tc.old,
		}
		old.Status.ReconciledGeneration = 1
		new := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-instance",
				Namespace:  "test-ns",
				Generation: 2,
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference: servicecatalog.PlanReference{
					ClusterServiceClassExternalName: clusterServiceClassExternalName,
					ClusterServicePlanExternalName:  clusterServicePlanExternalName,
				},
				ClusterServiceClassRef: &servicecatalog.ClusterObjectReference{},
				ClusterServicePlanRef:  &servicecatalog.ClusterObjectReference{},
			},
			Status: *tc.new,
		}
		new.Status.ReconciledGeneration = 1

		errs := ValidateServiceInstanceStatusUpdate(new, old)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
		if !tc.valid {
			for _, err := range errs {
				if !strings.Contains(err.Detail, tc.err) {
					t.Errorf("%v: Error %q did not contain expected message %q", tc.name, err.Detail, tc.err)
				}
			}
		}
	}

	for _, tc := range cases {
		old := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-instance",
				Namespace:  "test-ns",
				Generation: 2,
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference: servicecatalog.PlanReference{
					ServiceClassExternalName: serviceClassExternalName,
					ServicePlanExternalName:  servicePlanExternalName,
				},
				ServiceClassRef: &servicecatalog.LocalObjectReference{},
				ServicePlanRef:  &servicecatalog.LocalObjectReference{},
			},
			Status: *tc.old,
		}
		old.Status.ReconciledGeneration = 1
		new := &servicecatalog.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-instance",
				Namespace:  "test-ns",
				Generation: 2,
			},
			Spec: servicecatalog.ServiceInstanceSpec{
				PlanReference: servicecatalog.PlanReference{
					ClusterServiceClassExternalName: clusterServiceClassExternalName,
					ClusterServicePlanExternalName:  clusterServicePlanExternalName,
				},
				ClusterServiceClassRef: &servicecatalog.ClusterObjectReference{},
				ClusterServicePlanRef:  &servicecatalog.ClusterObjectReference{},
			},
			Status: *tc.new,
		}
		new.Status.ReconciledGeneration = 1

		errs := ValidateServiceInstanceStatusUpdate(new, old)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
		if !tc.valid {
			for _, err := range errs {
				if !strings.Contains(err.Detail, tc.err) {
					t.Errorf("%v: Error %q did not contain expected message %q", tc.name, err.Detail, tc.err)
				}
			}
		}
	}
}

func TestValidateServiceInstanceReferencesUpdate(t *testing.T) {
	cases := []struct {
		name  string
		old   *servicecatalog.ServiceInstance
		new   *servicecatalog.ServiceInstance
		valid bool
	}{
		{
			name: "valid clusterserviceclass and clusterserviceplan update",
			old: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServiceClassRef = nil
				i.Spec.ClusterServicePlanRef = nil
				return i
			}(),
			new:   validClusterRefServiceInstance(),
			valid: true,
		},
		{
			name: "valid serviceclass and serviceplan update",
			old: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Spec.ServiceClassRef = nil
				i.Spec.ServicePlanRef = nil
				return i
			}(),
			new:   validNamespacedRefServiceInstance(),
			valid: true,
		},
		{
			name: "invalid clusterserviceclass update",
			old:  validClusterRefServiceInstance(),
			new: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServiceClassRef = &servicecatalog.ClusterObjectReference{
					Name: "new-class-name",
				}
				return i
			}(),
			valid: false,
		},
		{
			name: "invalid serviceclass update",
			old:  validNamespacedRefServiceInstance(),
			new: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Spec.ServiceClassRef = &servicecatalog.LocalObjectReference{
					Name: "new-class-name",
				}
				return i
			}(),
			valid: false,
		},
		{
			name: "direct update to clusterserviceplan ref",
			old:  validClusterRefServiceInstance(),
			new: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServicePlanRef = &servicecatalog.ClusterObjectReference{
					Name: "new-plan-name",
				}
				return i
			}(),
			valid: false,
		},
		{
			name: "direct update to plan ref",
			old:  validNamespacedRefServiceInstance(),
			new: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Spec.ServicePlanRef = &servicecatalog.LocalObjectReference{
					Name: "new-plan-name",
				}
				return i
			}(),
			valid: false,
		},
		{
			name: "valid clusterserviceplan update from name change",
			old: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServicePlanRef = nil
				return i
			}(),
			new: func() *servicecatalog.ServiceInstance {
				i := validClusterRefServiceInstance()
				i.Spec.ClusterServicePlanRef = &servicecatalog.ClusterObjectReference{
					Name: "new-plan-name",
				}
				return i
			}(),
			valid: true,
		},
		{
			name: "valid serviceplan update from name change",
			old: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Spec.ServicePlanRef = nil
				return i
			}(),
			new: func() *servicecatalog.ServiceInstance {
				i := validNamespacedRefServiceInstance()
				i.Spec.ServicePlanRef = &servicecatalog.LocalObjectReference{
					Name: "new-plan-name",
				}
				return i
			}(),
			valid: true,
		},
		{
			name:  "in-progress operation",
			old:   validServiceInstanceWithInProgressProvision(),
			new:   validServiceInstanceWithInProgressProvision(),
			valid: false,
		},
	}

	for _, tc := range cases {
		errs := ValidateServiceInstanceReferencesUpdate(tc.new, tc.old)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}

func TestValidateClusterOrNamespacedPlanReference(t *testing.T) {
	cFields := []string{
		"ClusterServiceClassExternalName",
		"ClusterServiceClassExternalID",
		"ClusterServiceClassName",
	}
	pFields := []string{
		"ServiceClassExternalName",
		"ServiceClassExternalID",
		"ServiceClassName",
	}

	// Test permutations of cluster & plan fields set, these should never be valid
	cases := []servicecatalog.PlanReference{}
	for _, c := range cFields {
		for _, p := range pFields {
			pref := servicecatalog.PlanReference{}
			elem := reflect.ValueOf(&pref).Elem()
			elem.FieldByName(c).SetString("foo")
			elem.FieldByName(p).SetString("bar")
			cases = append(cases, pref)
		}
	}

	for _, testPlanRef := range cases {
		expectedErr := "instances can only refer to a cluster or namespaced class or plan type, but not both"
		errs := validatePlanReference(&testPlanRef, field.NewPath("spec"))
		if len(errs) == 0 {
			t.Fatalf(`Expected error "%s", but no error was found`, expectedErr)
		}

		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), expectedErr) {
				found = true
			}
		}
		if !found {
			t.Fatalf(`TestValidateClusterOrNamespacedPlanReference: did not find expected error "%s" in errors: %v`, expectedErr, errs)
		}
	}
}

func TestValidatePlanReference(t *testing.T) {
	cases := []struct {
		name          string
		ref           servicecatalog.PlanReference
		valid         bool
		expectedError string
	}{
		{
			name:          "invalid -- empty struct",
			ref:           servicecatalog.PlanReference{},
			valid:         false,
			expectedError: "plan references must have a class reference set",
		},
		{
			name:  "valid -- cluster service external names",
			ref:   validPlanReferenceClusterServiceExternalName(),
			valid: true,
		},
		{
			name:  "valid -- cluster service external ids",
			ref:   validPlanReferenceClusterServiceExternalID(),
			valid: true,
		},
		{
			name:  "valid -- cluster k8s",
			ref:   validPlanReferenceClusterK8S(),
			valid: true,
		},
		{
			name:  "valid -- service external names",
			ref:   validPlanReferenceServiceExternalName(),
			valid: true,
		},
		{
			name:  "valid -- service external ids",
			ref:   validPlanReferenceServiceExternalID(),
			valid: true,
		},
		{
			name:  "valid -- k8s",
			ref:   validPlanReferenceK8S(),
			valid: true,
		},
		{
			name: "invalid -- cluster external class name, k8s plan",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassExternalName: clusterServiceClassExternalName,
				ClusterServicePlanName:          clusterServicePlanExternalName,
			},
			valid:         false,
			expectedError: "must specify clusterServicePlanExternalName",
		},
		{
			name: "invalid -- cluster external class name, external plan id",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassExternalName: clusterServiceClassExternalName,
				ClusterServicePlanExternalID:    clusterServicePlanExternalID,
			},
			valid:         false,
			expectedError: "must specify clusterServicePlanExternalName",
		},
		{
			name: "invalid -- cluster external class id, k8s plan",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassExternalID: clusterServiceClassExternalID,
				ClusterServicePlanName:        clusterServicePlanName,
			},
			valid:         false,
			expectedError: "must specify clusterServicePlanExternalID",
		},
		{
			name: "invalid -- cluster k8s class, external plan name",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassName:        clusterServiceClassName,
				ClusterServicePlanExternalName: clusterServicePlanExternalName,
			},
			valid:         false,
			expectedError: "must specify clusterServicePlanName",
		},
		{
			name: "valid -- cluster k8s class, no plan",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassName: clusterServiceClassName,
			},
			valid: true,
		},
		{
			name: "invalid -- cluster k8s class, external plan id",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassName:      clusterServiceClassName,
				ClusterServicePlanExternalID: clusterServicePlanExternalID,
			},
			valid:         false,
			expectedError: "must specify clusterServicePlanName",
		},
		{
			name: "invalid -- cluster k8s class, external class name",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassName:         clusterServiceClassName,
				ClusterServiceClassExternalName: clusterServiceClassExternalName,
			},
			valid:         false,
			expectedError: "exactly one of clusterServiceClassExternalName",
		},
		{
			name: "invalid -- cluster k8s class, external class id",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassName:       clusterServiceClassName,
				ClusterServiceClassExternalID: clusterServiceClassExternalID,
			},
			valid:         false,
			expectedError: "exactly one of clusterServiceClassExternalName",
		},
		{
			name: "invalid -- cluster external class name, external class id",
			ref: servicecatalog.PlanReference{
				ClusterServiceClassExternalName: clusterServiceClassExternalName,
				ClusterServiceClassExternalID:   clusterServiceClassExternalID,
			},
			valid:         false,
			expectedError: "exactly one of clusterServiceClassExternalName",
		},
		{
			name: "invalid -- cluster k8s plan, external plan name",
			ref: servicecatalog.PlanReference{
				ClusterServicePlanName:         clusterServicePlanName,
				ClusterServicePlanExternalName: clusterServicePlanExternalName,
			},
			valid:         false,
			expectedError: "exactly one of clusterServicePlanExternalName",
		},
		{
			name: "invalid -- cluster k8s plan, external plan id",
			ref: servicecatalog.PlanReference{
				ClusterServicePlanName:       clusterServicePlanName,
				ClusterServicePlanExternalID: clusterServicePlanExternalID,
			},
			valid:         false,
			expectedError: "exactly one of clusterServicePlanExternalName",
		},
		{
			name: "invalid -- cluster external plan name, external plan id",
			ref: servicecatalog.PlanReference{
				ClusterServicePlanExternalName: clusterServicePlanExternalName,
				ClusterServicePlanExternalID:   clusterServicePlanExternalID,
			},
			valid:         false,
			expectedError: "exactly one of clusterServicePlanExternalName",
		},
		{
			name: "invalid -- external class name, k8s plan",
			ref: servicecatalog.PlanReference{
				ServiceClassExternalName: serviceClassExternalName,
				ServicePlanName:          servicePlanExternalName,
			},
			valid:         false,
			expectedError: "must specify servicePlanExternalName",
		},
		{
			name: "invalid -- external class name, external plan id",
			ref: servicecatalog.PlanReference{
				ServiceClassExternalName: serviceClassExternalName,
				ServicePlanExternalID:    servicePlanExternalID,
			},
			valid:         false,
			expectedError: "must specify servicePlanExternalName",
		},
		{
			name: "invalid -- external class id, k8s plan",
			ref: servicecatalog.PlanReference{
				ServiceClassExternalID: serviceClassExternalID,
				ServicePlanName:        servicePlanName,
			},
			valid:         false,
			expectedError: "must specify servicePlanExternalID",
		},
		{
			name: "invalid -- k8s class, external plan name",
			ref: servicecatalog.PlanReference{
				ServiceClassName:        serviceClassName,
				ServicePlanExternalName: servicePlanExternalName,
			},
			valid:         false,
			expectedError: "must specify servicePlanName",
		},
		{
			name: "invalid -- k8s class, external plan id",
			ref: servicecatalog.PlanReference{
				ServiceClassName:      serviceClassName,
				ServicePlanExternalID: servicePlanExternalID,
			},
			valid:         false,
			expectedError: "must specify servicePlanName",
		},
		{
			name: "invalid -- k8s class, external class name",
			ref: servicecatalog.PlanReference{
				ServiceClassName:         serviceClassName,
				ServiceClassExternalName: serviceClassExternalName,
			},
			valid:         false,
			expectedError: "exactly one of serviceClassExternalName",
		},
		{
			name: "invalid -- k8s class, external class id",
			ref: servicecatalog.PlanReference{
				ServiceClassName:       serviceClassName,
				ServiceClassExternalID: serviceClassExternalID,
			},
			valid:         false,
			expectedError: "exactly one of serviceClassExternalName",
		},
		{
			name: "invalid -- external class name, external class id",
			ref: servicecatalog.PlanReference{
				ServiceClassExternalName: serviceClassExternalName,
				ServiceClassExternalID:   serviceClassExternalID,
			},
			valid:         false,
			expectedError: "exactly one of serviceClassExternalName",
		},
		{
			name: "invalid -- k8s plan, external plan name",
			ref: servicecatalog.PlanReference{
				ServicePlanName:         servicePlanName,
				ServicePlanExternalName: servicePlanExternalName,
			},
			valid:         false,
			expectedError: "exactly one of servicePlanExternalName",
		},
		{
			name: "invalid -- k8s plan, external plan id",
			ref: servicecatalog.PlanReference{
				ServicePlanName:       servicePlanName,
				ServicePlanExternalID: servicePlanExternalID,
			},
			valid:         false,
			expectedError: "exactly one of servicePlanExternalName",
		},
		{
			name: "invalid -- external plan name, external plan id",
			ref: servicecatalog.PlanReference{
				ServicePlanExternalName: servicePlanExternalName,
				ServicePlanExternalID:   servicePlanExternalID,
			},
			valid:         false,
			expectedError: "exactly one of servicePlanExternalName",
		},
	}
	for _, tc := range cases {
		errs := validatePlanReference(&tc.ref, field.NewPath("spec"))
		if len(errs) != 0 {
			if tc.valid {
				t.Errorf("%v: unexpected error: %v", tc.name, errs)
				continue
			}
			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), tc.expectedError) {
					found = true
				}
			}
			if !found {
				t.Errorf("%v: did not find expected error %q in errors: %v", tc.name, tc.expectedError, errs)
				continue
			}
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}

func TestValidatePlanReferenceUpdate(t *testing.T) {
	cases := []struct {
		name          string
		old           servicecatalog.PlanReference
		new           servicecatalog.PlanReference
		valid         bool
		expectedError string
	}{
		{
			name:  "valid -- no changes external",
			old:   validPlanReferenceClusterServiceExternalName(),
			new:   validPlanReferenceClusterServiceExternalName(),
			valid: true,
		},
		{
			name:  "valid -- no changes k8s",
			old:   validPlanReferenceClusterK8S(),
			new:   validPlanReferenceClusterK8S(),
			valid: true,
		},
		{
			name: "invalid -- changing external class name",
			old:  validPlanReferenceClusterServiceExternalName(),
			new: servicecatalog.PlanReference{
				ClusterServiceClassExternalName: "new-class",
				ClusterServicePlanExternalName:  clusterServicePlanExternalName,
			},
			valid:         false,
			expectedError: "clusterServiceClassExternalName",
		},
		{
			name: "valid -- changing external plan name",
			old:  validPlanReferenceClusterServiceExternalName(),
			new: servicecatalog.PlanReference{
				ClusterServiceClassExternalName: clusterServiceClassExternalName,
				ClusterServicePlanExternalName:  "new-plan",
			},
			valid: true,
		},
		{
			name: "invalid -- changing external class id",
			old:  validPlanReferenceClusterServiceExternalID(),
			new: servicecatalog.PlanReference{
				ClusterServiceClassExternalID: "new-class",
				ClusterServicePlanExternalID:  clusterServicePlanExternalID,
			},
			valid:         false,
			expectedError: "clusterServiceClassExternalID",
		},
		{
			name: "valid -- changing external plan id",
			old:  validPlanReferenceClusterServiceExternalID(),
			new: servicecatalog.PlanReference{
				ClusterServiceClassExternalID: clusterServiceClassExternalID,
				ClusterServicePlanExternalID:  "new-plan",
			},
			valid: true,
		},
		{
			name: "invalid -- changing k8s class name",
			old:  validPlanReferenceClusterK8S(),
			new: servicecatalog.PlanReference{
				ClusterServiceClassName: "new-class",
				ClusterServicePlanName:  clusterServicePlanName,
			},
			valid:         false,
			expectedError: "clusterServiceClassName",
		},
		{
			name: "valid -- changing k8s plan name",
			old:  validPlanReferenceClusterK8S(),
			new: servicecatalog.PlanReference{
				ClusterServiceClassName: clusterServiceClassName,
				ClusterServicePlanName:  "new-plan",
			},
			valid: true,
		},
	}
	for _, tc := range cases {
		errs := validatePlanReferenceUpdate(&tc.old, &tc.new, field.NewPath("spec"))
		if len(errs) != 0 {
			if tc.valid {
				t.Errorf("%v: unexpected error: %v", tc.name, errs)
				continue
			}
			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), tc.expectedError) {
					found = true
				}
			}
			if !found {
				t.Errorf("%v: did not find expected error %q in errors: %v", tc.name, tc.expectedError, errs)
				continue
			}
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}
