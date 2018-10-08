/*
Copyright 2018 The Kubernetes Authors.

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

package class

import (
	"bytes"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/internal/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	svcatfake "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestDescribeCommand(t *testing.T) {
	const namespace = "default"
	testcases := []struct {
		name            string // Test Nmae
		cmd             string // Command to run
		golden          string // Relative path to a golden file, compared to the command output
		continueOnError bool   // Should the test stop immediately if the command fails or continue and capture the console output
	}{
		{
			name:   "describe class by name",
			cmd:    "describe class user-provided-service",
			golden: "describe-class.txt",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup fake data for the app
			k8sClient := k8sfake.NewSimpleClientset()
			fakeClass := &v1beta1.ClusterServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name:            "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
					UID:             "7b3c2fe0-f711-11e7-aa44-0242ac110005",
					ResourceVersion: "3",
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					ClusterServiceBrokerName: "ups-broker",
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName:       "user-provided-service",
						ExternalID:         "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
						Description:        "A user provided service",
						Bindable:           true,
						BindingRetrievable: false,
						PlanUpdatable:      true,
					},
				},
				Status: v1beta1.ClusterServiceClassStatus{
					CommonServiceClassStatus: v1beta1.CommonServiceClassStatus{
						RemovedFromBrokerCatalog: false,
					},
				},
			}

			fakePlan1 := &v1beta1.ClusterServicePlan{
				ObjectMeta: v1.ObjectMeta{
					Name:            "86064792-7ea2-467b-af93-ac9694d96d52",
					UID:             "69ce3c3d-f7de-11e7-9c07-0242ac110006",
					ResourceVersion: "5",
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					ClusterServiceBrokerName: "ups-broker",
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: "default",
						ExternalID:   "86064792-7ea2-467b-af93-ac9694d96d52",
						Description:  "Sample plan description",
						Free:         true,
					},
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
					},
				},
				Status: v1beta1.ClusterServicePlanStatus{
					CommonServicePlanStatus: v1beta1.CommonServicePlanStatus{
						RemovedFromBrokerCatalog: false,
					},
				},
			}

			fakePlan2 := &v1beta1.ClusterServicePlan{
				ObjectMeta: v1.ObjectMeta{
					Name:            "cc0d7529-18e8-416d-8946-6f7456acd589",
					UID:             "7b497b48-f711-11e7-aa44-0242ac110005",
					ResourceVersion: "5",
				},
				Spec: v1beta1.ClusterServicePlanSpec{
					ClusterServiceBrokerName: "ups-broker",
					CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
						ExternalName: "premium",
						ExternalID:   "cc0d7529-18e8-416d-8946-6f7456acd589",
						Description:  "Premium plan",
						Free:         true,
					},
					ClusterServiceClassRef: v1beta1.ClusterObjectReference{
						Name: "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
					},
				},
				Status: v1beta1.ClusterServicePlanStatus{
					CommonServicePlanStatus: v1beta1.CommonServicePlanStatus{
						RemovedFromBrokerCatalog: false,
					},
				},
			}

			svcatClient := svcatfake.NewSimpleClientset(fakeClass, fakePlan1, fakePlan2)
			fakeApp, _ := svcat.NewApp(k8sClient, svcatClient, namespace)
			output := &bytes.Buffer{}
			cxt := svcattest.NewContext(output, fakeApp)

			// Initialize the command arguments
			cmd := &describeCmd{
				Context: cxt,
			}
			// Capture all output: stderr and stdout
			cmd.Context.Output = output

			cmd.Run()
			test.AssertEqualsGoldenFile(t, tc.golden, output.String())
		})
	}
}
