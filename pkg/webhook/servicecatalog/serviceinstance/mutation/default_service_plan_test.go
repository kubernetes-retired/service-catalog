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

package mutation_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/serviceinstance/mutation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubernetes-sigs/service-catalog/pkg/util"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhookutil"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestErrorWhenNoClassesSpecified(t *testing.T) {
	dsp := mutation.DefaultServicePlan{}

	mutateErr := dsp.SetDefaultPlan(context.Background(), newServiceInstance("dummy"), webhookutil.NewTracedLogger(uuid.NewUUID()))
	assertMutateError(t, mutateErr, "class not specified on ServiceInstance, cannot choose default plan", http.StatusInternalServerError)
}

func TestClusterServiceClassSpecified(t *testing.T) {
	const className = "csc"

	for tn, tc := range map[string]struct {
		instance *sc.ServiceInstance
		objects  []runtime.Object
		err      *webhookutil.WebhookError
	}{
		"SuccessWithClusterServiceClassName": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ClusterServiceClassName: className,
					},
				},
			},
			objects: []runtime.Object{
				newClusterServiceClass(className, className),
				newClusterServicePlans(className, 1, false)[0],
			},
		},
		"SuccessWithClusterServiceClassByField": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ClusterServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newClusterServiceClass(className, className),
				newClusterServicePlans(className, 1, false)[0],
			},
		},
		"SuccessWithManyPlansToDifferentClasses": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ClusterServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newClusterServiceClass(className, className),
				newClusterServicePlans(className, 2, true)[0],
				newClusterServicePlans(className, 2, true)[1],
			},
		},
		"ErrorWhenNoPlansSpecified": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ClusterServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newClusterServiceClass(className, className),
			},
			err: webhookutil.NewWebhookError(fmt.Sprintf("no ClusterServicePlans found at all for ClusterServiceClass %q", className), http.StatusForbidden),
		},
		"ErrorWhenMoreThenOnePlanSpecified": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ClusterServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newClusterServiceClass(className, className),
				newClusterServicePlans(className, 2, false)[0],
				newClusterServicePlans(className, 2, false)[1],
			},
			err: webhookutil.NewWebhookError(fmt.Sprintf("ClusterServiceClass (K8S: %v ExternalName: %v) has more than one plan, PlanName must be specified", className, className), http.StatusForbidden),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fakeClient := fake.NewFakeClientWithScheme(newTestScheme(t), tc.objects...)
			traced := webhookutil.NewTracedLogger(uuid.NewUUID())

			dsp := mutation.DefaultServicePlan{}
			dsp.InjectClient(fakeClient)

			mutateErr := dsp.SetDefaultPlan(context.Background(), tc.instance, traced)

			if tc.err != nil {
				assertMutateError(t, mutateErr, tc.err.Error(), tc.err.Code())
			} else {
				assert.Nil(t, mutateErr)
			}
		})
	}
}

func TestServiceClassSpecified(t *testing.T) {
	const className = "sc"
	const namespace = "dummy"

	for tn, tc := range map[string]struct {
		instance *sc.ServiceInstance
		objects  []runtime.Object
		err      *webhookutil.WebhookError
	}{
		"SuccessWithServiceClassName": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: namespace},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ServiceClassName: className,
					},
				},
			},
			objects: []runtime.Object{
				newServiceClass(className, className, namespace),
				newServicePlans(className, namespace, 1, false)[0],
			},
		},
		"SuccessWithServiceClassByField": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newServiceClass(className, className, namespace),
				newServicePlans(className, namespace, 1, false)[0],
			},
		},
		"SuccessWithManyPlansToDifferentClasses": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newServiceClass(className, className, namespace),
				newServicePlans(className, namespace, 2, true)[0],
				newServicePlans(className, namespace, 2, true)[1],
			},
		},
		"SuccessWithFindingDefaultPlanForManyServiceClasses": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newServiceClass(className, className, namespace),
				newServiceClass(className, className, "otherNamespace"),
				newServicePlans(className, namespace, 1, true)[0],
				newServicePlans(className, "otherNamespace", 1, true)[0],
			},
		},
		"ErrorWhenNoPlansSpecified": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newServiceClass(className, className, namespace),
			},
			err: webhookutil.NewWebhookError(fmt.Sprintf("no ServicePlans found at all for ServiceClass %q", className), http.StatusForbidden),
		},
		"ErrorWhenMoreThenOnePlanSpecified": {
			instance: &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "dummy"},
				Spec: sc.ServiceInstanceSpec{
					PlanReference: sc.PlanReference{
						ServiceClassExternalName: className,
					},
				},
			},
			objects: []runtime.Object{
				newServiceClass(className, className, namespace),
				newServicePlans(className, namespace, 2, false)[0],
				newServicePlans(className, namespace, 2, false)[1],
			},
			err: webhookutil.NewWebhookError(fmt.Sprintf("ServiceClass (K8S: %v ExternalName: %v) has more than one plan, PlanName must be specified", className, className), http.StatusForbidden),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			fakeClient := fake.NewFakeClientWithScheme(newTestScheme(t), tc.objects...)
			traced := webhookutil.NewTracedLogger(uuid.NewUUID())

			dsp := mutation.DefaultServicePlan{}
			dsp.InjectClient(fakeClient)

			mutateErr := dsp.SetDefaultPlan(context.Background(), tc.instance, traced)

			if tc.err != nil {
				assertMutateError(t, mutateErr, tc.err.Error(), tc.err.Code())
			} else {
				assert.Nil(t, mutateErr)
			}
		})
	}
}

func newTestScheme(t *testing.T) *runtime.Scheme {
	sch, err := sc.SchemeBuilderRuntime.Build()
	require.NoError(t, err)
	sc.AddToScheme(scheme.Scheme)

	return sch
}

// newServiceInstance returns a new instance for the specified namespace.
func newServiceInstance(namespace string) *sc.ServiceInstance {
	return &sc.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: namespace},
	}
}

// newClusterServiceClass returns a new serviceclass.
func newClusterServiceClass(id string, name string) *sc.ClusterServiceClass {
	sc := &sc.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
			Labels: map[string]string{
				sc.GroupName + "/" + sc.FilterSpecExternalID:   util.GenerateSHA(id),
				sc.GroupName + "/" + sc.FilterSpecExternalName: util.GenerateSHA(name),
			},
		},
		Spec: sc.ClusterServiceClassSpec{
			CommonServiceClassSpec: sc.CommonServiceClassSpec{
				ExternalID:   id,
				ExternalName: name,
			},
		},
	}
	return sc
}

// newServiceClass returns a new serviceclass.
func newServiceClass(id string, name string, namespace string) *sc.ServiceClass {
	sc := &sc.ServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: namespace,
			Labels: map[string]string{
				sc.GroupName + "/" + sc.FilterSpecExternalID:   util.GenerateSHA(id),
				sc.GroupName + "/" + sc.FilterSpecExternalName: util.GenerateSHA(name),
			},
		},
		Spec: sc.ServiceClassSpec{
			CommonServiceClassSpec: sc.CommonServiceClassSpec{
				ExternalID:   id,
				ExternalName: name,
			},
		},
	}
	return sc
}

// newClusterServicePlans returns new serviceplans.
func newClusterServicePlans(classname string, count uint, useDifferentClasses bool) []*sc.ClusterServicePlan {
	sp1 := &sc.ClusterServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar-id",
			Labels: map[string]string{
				sc.GroupName + "/" + sc.FilterSpecExternalID:                 util.GenerateSHA("12345"),
				sc.GroupName + "/" + sc.FilterSpecExternalName:               util.GenerateSHA("bar"),
				sc.GroupName + "/" + sc.FilterSpecClusterServiceClassRefName: util.GenerateSHA(classname),
			},
		},
		Spec: sc.ClusterServicePlanSpec{
			CommonServicePlanSpec: sc.CommonServicePlanSpec{
				ExternalName: "bar",
				ExternalID:   "12345",
			},
			ClusterServiceClassRef: sc.ClusterObjectReference{
				Name: classname,
			},
		},
	}
	if useDifferentClasses {
		classname = "different-serviceclass"
	}
	sp2 := &sc.ClusterServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name: "baz-id",
			Labels: map[string]string{
				sc.GroupName + "/" + sc.FilterSpecExternalID:                 util.GenerateSHA("23456"),
				sc.GroupName + "/" + sc.FilterSpecExternalName:               util.GenerateSHA("baz"),
				sc.GroupName + "/" + sc.FilterSpecClusterServiceClassRefName: util.GenerateSHA(classname),
			},
		},
		Spec: sc.ClusterServicePlanSpec{
			CommonServicePlanSpec: sc.CommonServicePlanSpec{
				ExternalName: "baz",
				ExternalID:   "23456",
			},
			ClusterServiceClassRef: sc.ClusterObjectReference{
				Name: classname,
			},
		},
	}

	if 0 == count {
		return []*sc.ClusterServicePlan{}
	}
	if 1 == count {
		return []*sc.ClusterServicePlan{sp1}
	}
	if 2 == count {
		return []*sc.ClusterServicePlan{sp1, sp2}
	}
	return []*sc.ClusterServicePlan{}
}

// newServicePlans returns new serviceplans.
func newServicePlans(classname string, namespace string, count uint, useDifferentClasses bool) []*sc.ServicePlan {
	sp1 := &sc.ServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar-id",
			Namespace: namespace,
			Labels: map[string]string{
				sc.GroupName + "/" + sc.FilterSpecExternalID:          util.GenerateSHA("12345"),
				sc.GroupName + "/" + sc.FilterSpecExternalName:        util.GenerateSHA("bar"),
				sc.GroupName + "/" + sc.FilterSpecServiceClassRefName: util.GenerateSHA(classname),
			},
		},
		Spec: sc.ServicePlanSpec{
			CommonServicePlanSpec: sc.CommonServicePlanSpec{
				ExternalName: "bar",
				ExternalID:   "12345",
			},
			ServiceClassRef: sc.LocalObjectReference{
				Name: classname,
			},
		},
	}
	if useDifferentClasses {
		classname = "different-serviceclass"
	}
	sp2 := &sc.ServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "baz-id",
			Namespace: namespace,
			Labels: map[string]string{
				sc.GroupName + "/" + sc.FilterSpecExternalID:          util.GenerateSHA("23456"),
				sc.GroupName + "/" + sc.FilterSpecExternalName:        util.GenerateSHA("bar"),
				sc.GroupName + "/" + sc.FilterSpecServiceClassRefName: util.GenerateSHA(classname),
			},
		},
		Spec: sc.ServicePlanSpec{
			CommonServicePlanSpec: sc.CommonServicePlanSpec{
				ExternalName: "baz",
				ExternalID:   "23456",
			},
			ServiceClassRef: sc.LocalObjectReference{
				Name: classname,
			},
		},
	}

	if 0 == count {
		return []*sc.ServicePlan{}
	}
	if 1 == count {
		return []*sc.ServicePlan{sp1}
	}
	if 2 == count {
		return []*sc.ServicePlan{sp1, sp2}
	}
	return []*sc.ServicePlan{}
}

func assertMutateError(t *testing.T, actualErr *webhookutil.WebhookError, expMsg string, expCode int32) {
	assert.Equal(t, expMsg, actualErr.Error())
	assert.Equal(t, expCode, actualErr.Code())
}
