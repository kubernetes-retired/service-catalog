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
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

func TestValidateClusterServiceBroker(t *testing.T) {
	cases := []struct {
		name   string
		broker *servicecatalog.ClusterServiceBroker
		valid  bool
	}{
		{
			// covers the case where there is no AuthInfo field specified. the validator should
			// ignore the field and still succeed the validation
			name: "valid clusterservicebroker - no auth secret",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker - basic auth - secret",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					AuthInfo: &servicecatalog.ClusterServiceBrokerAuthInfo{
						Basic: &servicecatalog.ClusterBasicAuthConfig{
							SecretRef: &servicecatalog.ObjectReference{
								Namespace: "test-ns",
								Name:      "test-secret",
							},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						URL:            "http://example.com",
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker - bearer auth - secret",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					AuthInfo: &servicecatalog.ClusterServiceBrokerAuthInfo{
						Bearer: &servicecatalog.ClusterBearerTokenAuthConfig{
							SecretRef: &servicecatalog.ObjectReference{
								Namespace: "test-ns",
								Name:      "test-secret",
							},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid clusterservicebroker - clusterservicebroker with namespace",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "oops",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid clusterservicebroker - basic auth - secret missing namespace",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					AuthInfo: &servicecatalog.ClusterServiceBrokerAuthInfo{
						Basic: &servicecatalog.ClusterBasicAuthConfig{
							SecretRef: &servicecatalog.ObjectReference{
								Name: "test-secret",
							},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid clusterservicebroker - basic auth - secret missing name",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					AuthInfo: &servicecatalog.ClusterServiceBrokerAuthInfo{
						Basic: &servicecatalog.ClusterBasicAuthConfig{
							SecretRef: &servicecatalog.ObjectReference{
								Namespace: "test-ns",
							},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid clusterservicebroker - bearer auth - secret missing namespace",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					AuthInfo: &servicecatalog.ClusterServiceBrokerAuthInfo{
						Bearer: &servicecatalog.ClusterBearerTokenAuthConfig{
							SecretRef: &servicecatalog.ObjectReference{
								Name: "test-secret",
							},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid clusterservicebroker - bearer auth - secret missing name",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					AuthInfo: &servicecatalog.ClusterServiceBrokerAuthInfo{
						Bearer: &servicecatalog.ClusterBearerTokenAuthConfig{
							SecretRef: &servicecatalog.ObjectReference{
								Namespace: "test-ns",
							},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid clusterservicebroker - CABundle present with InsecureSkipTLSVerify",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL: "http://example.com",
						InsecureSkipTLSVerify: true,
						CABundle:              []byte("fake CABundle"),
						RelistBehavior:        servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration:        &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "valid clusterservicebroker - InsecureSkipTLSVerify without CABundle",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL: "http://example.com",
						InsecureSkipTLSVerify: true,
						RelistBehavior:        servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration:        &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker - CABundle without InsecureSkipTLSVerify",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						CABundle:       []byte("fake CABundle"),
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker - manual behavior with RelistDuration",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker - manual behavior without RelistDuration",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						RelistDuration: nil,
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker - duration behavior defaulting to controller provided value",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: nil,
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid clusterservicebroker - relistBehavior is invalid",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: "Junk",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid clusterservicebroker - relistBehavior is empty",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: "",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid clusterservicebroker - negative relistRequests value",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: -1,
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid clusterservicebroker - negative relistDuration value",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: -15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "valid clusterservicebroker - catalogRequirements.serviceClass",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-broker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						CatalogRestrictions: &servicecatalog.CatalogRestrictions{
							ServiceClass: []string{
								"name==foobar",
							},
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker - complex catalogRequirements.serviceClass",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-broker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						CatalogRestrictions: &servicecatalog.CatalogRestrictions{
							ServiceClass: []string{
								"name==foobar",
								"externalName in (foobar, bazboof, wizzbang)",
							},
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid clusterservicebroker - catalogRequirements.serviceClass",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-broker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						CatalogRestrictions: &servicecatalog.CatalogRestrictions{
							ServiceClass: []string{
								"invalid restriction",
							},
						},
					},
				},
			},
			valid: false,
		},
		{
			name: "valid clusterservicebroker - catalogRequirements.servicePlan",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-broker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						CatalogRestrictions: &servicecatalog.CatalogRestrictions{
							ServicePlan: []string{
								"name==foobar",
							},
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker - complex catalogRequirements.servicePlan",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-broker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						CatalogRestrictions: &servicecatalog.CatalogRestrictions{
							ServicePlan: []string{
								"name==foobar",
								"externalName in (foobar, bazboof, wizzbang)",
							},
						},
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid clusterservicebroker - catalogRequirements.servicePlan",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-broker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						CatalogRestrictions: &servicecatalog.CatalogRestrictions{
							ServicePlan: []string{
								"invalid restriction",
							},
						},
					},
				},
			},
			valid: false,
		},
		{
			name: "valid clusterservicebroker - catalogRequirements with serviceClass and servicePlan",
			broker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-broker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						CatalogRestrictions: &servicecatalog.CatalogRestrictions{
							ServiceClass: []string{
								"name==barfoobar",
								"externalName in (barfoobar, batbazboof, batwizzbang)",
							},
							ServicePlan: []string{
								"name==foobar",
								"externalName in (foobar, bazboof, wizzbang)",
							},
						},
					},
				},
			},
			valid: true,
		},
	}

	for _, tc := range cases {
		errs := ValidateClusterServiceBroker(tc.broker)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			return
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}

	updateCases := []struct {
		name      string
		newBroker *servicecatalog.ClusterServiceBroker
		oldBroker *servicecatalog.ClusterServiceBroker
		valid     bool
	}{
		{
			name: "valid clusterservicebroker update - equal relistRequests value",
			newBroker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 1,
					},
				},
			},
			oldBroker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 1,
					},
				},
			},
			valid: true,
		},
		{
			name: "valid clusterservicebroker update - increasing relistRequests value",
			newBroker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 2,
					},
				},
			},
			oldBroker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 1,
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid clusterservicebroker update - nonincreasing relistRequests value",
			newBroker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 1,
					},
				},
			},
			oldBroker: &servicecatalog.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 2,
					},
				},
			},
			valid: false,
		},
	}
	for _, tc := range updateCases {
		errs := ValidateClusterServiceBrokerUpdate(tc.newBroker, tc.oldBroker)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}

func TestValidateServiceBroker(t *testing.T) {
	cases := []struct {
		name   string
		broker *servicecatalog.ServiceBroker
		valid  bool
	}{
		{
			// covers the case where there is no AuthInfo field specified. the validator should
			// ignore the field and still succeed the validation
			name: "valid servicebroker - no auth secret",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid servicebroker - basic auth - secret",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					AuthInfo: &servicecatalog.ServiceBrokerAuthInfo{
						Basic: &servicecatalog.BasicAuthConfig{
							SecretRef: &servicecatalog.LocalObjectReference{
								Name: "test-secret",
							},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						URL:            "http://example.com",
					},
				},
			},
			valid: true,
		},
		{
			name: "valid servicebroker - bearer auth - secret",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					AuthInfo: &servicecatalog.ServiceBrokerAuthInfo{
						Bearer: &servicecatalog.BearerTokenAuthConfig{
							SecretRef: &servicecatalog.LocalObjectReference{
								Name: "test-secret",
							},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid servicebroker - servicebroker without namespace",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-clusterservicebroker",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid servicebroker - basic auth - secret missing name",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					AuthInfo: &servicecatalog.ServiceBrokerAuthInfo{
						Basic: &servicecatalog.BasicAuthConfig{
							SecretRef: &servicecatalog.LocalObjectReference{},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid servicebroker - bearer auth - secret missing name",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					AuthInfo: &servicecatalog.ServiceBrokerAuthInfo{
						Bearer: &servicecatalog.BearerTokenAuthConfig{
							SecretRef: &servicecatalog.LocalObjectReference{},
						},
					},
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid servicebroker - CABundle present with InsecureSkipTLSVerify",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL: "http://example.com",
						InsecureSkipTLSVerify: true,
						CABundle:              []byte("fake CABundle"),
						RelistBehavior:        servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration:        &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: false,
		},
		{
			name: "valid servicebroker - InsecureSkipTLSVerify without CABundle",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL: "http://example.com",
						InsecureSkipTLSVerify: true,
						RelistBehavior:        servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration:        &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid servicebroker - CABundle without InsecureSkipTLSVerify",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						CABundle:       []byte("fake CABundle"),
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid servicebroker - manual behavior with RelistDuration",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
					},
				},
			},
			valid: true,
		},
		{
			name: "valid servicebroker - manual behavior without RelistDuration",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorManual,
						RelistDuration: nil,
					},
				},
			},
			valid: true,
		},
		{
			name: "valid servicebroker - duration behavior defaulting to controller provided value",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: nil,
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid servicebroker - relistBehavior is invalid",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: "Junk",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid servicebroker - relistBehavior is empty",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: "",
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid servicebroker - negative relistRequests value",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: -1,
					},
				},
			},
			valid: false,
		},
		{
			name: "invalid servicebroker - negative relistDuration value",
			broker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: -15 * time.Minute},
					},
				},
			},
			valid: false,
		},
	}

	for _, tc := range cases {
		errs := ValidateServiceBroker(tc.broker)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}

	updateCases := []struct {
		name      string
		newBroker *servicecatalog.ServiceBroker
		oldBroker *servicecatalog.ServiceBroker
		valid     bool
	}{
		{
			name: "valid servicebroker update - equal relistRequests value",
			newBroker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 1,
					},
				},
			},
			oldBroker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 1,
					},
				},
			},
			valid: true,
		},
		{
			name: "valid servicebroker update - increasing relistRequests value",
			newBroker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 2,
					},
				},
			},
			oldBroker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 1,
					},
				},
			},
			valid: true,
		},
		{
			name: "invalid servicebroker update - nonincreasing relistRequests value",
			newBroker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 1,
					},
				},
			},
			oldBroker: &servicecatalog.ServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusterservicebroker",
					Namespace: "test-ns",
				},
				Spec: servicecatalog.ServiceBrokerSpec{
					CommonServiceBrokerSpec: servicecatalog.CommonServiceBrokerSpec{
						URL:            "http://example.com",
						RelistBehavior: servicecatalog.ServiceBrokerRelistBehaviorDuration,
						RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
						RelistRequests: 2,
					},
				},
			},
			valid: false,
		},
	}
	for _, tc := range updateCases {
		errs := ValidateServiceBrokerUpdate(tc.newBroker, tc.oldBroker)
		if len(errs) != 0 && tc.valid {
			t.Errorf("%v: unexpected error: %v", tc.name, errs)
			continue
		} else if len(errs) == 0 && !tc.valid {
			t.Errorf("%v: unexpected success", tc.name)
		}
	}
}
