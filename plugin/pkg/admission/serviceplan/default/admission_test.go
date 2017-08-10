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

package defaultserviceplan

import (
	//	"fmt"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/admission"
	//	"k8s.io/client-go/pkg/api/v1"
	core "k8s.io/client-go/testing"

	informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/internalversion"
	//	internalversion "github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/servicecatalog/internalversion"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/internalclientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/internalclientset/fake"

	scadmission "github.com/kubernetes-incubator/service-catalog/pkg/apiserver/admission"
)

// newHandlerForTest returns a configured handler for testing.
func newHandlerForTest(internalClient internalclientset.Interface) (admission.Interface, informers.SharedInformerFactory, error) {
	f := informers.NewSharedInformerFactory(internalClient, 5*time.Minute)
	handler, err := NewDefaultServicePlan()
	if err != nil {
		return nil, f, err
	}
	pluginInitializer := scadmission.NewPluginInitializer(internalClient, f, nil, nil)
	pluginInitializer.Initialize(handler)
	err = admission.Validate(handler)
	return handler, f, err
}

// newMockServiceCatalogClientForTest creates a mock client that returns a client
// configured for the specified list of namespaces with the specified phase.
func newMockServiceCatalogClientForTest(sc *servicecatalog.ServiceClass) *fake.Clientset {
	mockClient := &fake.Clientset{}
	mockClient.AddReactor("get", "serviceclasses", func(action core.Action) (bool, runtime.Object, error) {
		return true, sc, nil
	})
	return mockClient
}

// newMockClientForTest creates a mock client.
func newMockClientForTest() *fake.Clientset {
	mockClient := &fake.Clientset{}
	return mockClient
}

// newBroker returns a new broker for testing.
func newBroker() servicecatalog.Broker {
	return servicecatalog.Broker{
		ObjectMeta: metav1.ObjectMeta{Name: "broker"},
	}
}

// newInstance returns a new instance for the specified namespace.
func newInstance(namespace string) servicecatalog.Instance {
	return servicecatalog.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: namespace},
	}
}

// newServiceClass returns a new instance with the specified plans.
func newServiceClass(name string, plans ...string) *servicecatalog.ServiceClass {
	sc := &servicecatalog.ServiceClass{ObjectMeta: metav1.ObjectMeta{Name: name}}
	for _, plan := range plans {
		sc.Plans = append(sc.Plans, servicecatalog.ServicePlan{Name: plan})
	}
	return sc
}

func TestWithPlanWorks(t *testing.T) {
	mockSCClient := newMockServiceCatalogClientForTest(&servicecatalog.ServiceClass{})
	handler, informerFactory, err := newHandlerForTest(mockSCClient)
	if err != nil {
		t.Errorf("unexpected error initializing handler: %v", err)
	}
	informerFactory.Start(wait.NeverStop)

	instance := newInstance("dummy")
	instance.Spec.ServiceClassName = "foo"
	instance.Spec.PlanName = "bar"

	err = handler.Admit(admission.NewAttributesRecord(&instance, nil, servicecatalog.Kind("Instance").WithVersion("version"), instance.Namespace, instance.Name, servicecatalog.Resource("instances").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		actions := ""
		for _, action := range mockSCClient.Actions() {
			actions = actions + action.GetVerb() + ":" + action.GetResource().Resource + ":" + action.GetSubresource() + ", "
		}
		t.Errorf("unexpected error returned from admission handler: %v", actions)
	}
}

func TestWithNoPlanFailsWithNoServiceClass(t *testing.T) {
	mockSCClient := newMockServiceCatalogClientForTest(&servicecatalog.ServiceClass{})
	handler, informerFactory, err := newHandlerForTest(mockSCClient)
	if err != nil {
		t.Errorf("unexpected error initializing handler: %v", err)
	}
	informerFactory.Start(wait.NeverStop)

	instance := newInstance("dummy")
	instance.Spec.ServiceClassName = "foo"

	err = handler.Admit(admission.NewAttributesRecord(&instance, nil, servicecatalog.Kind("Instance").WithVersion("version"), instance.Namespace, instance.Name, servicecatalog.Resource("instances").WithVersion("version"), "", admission.Create, nil))
	if err == nil {
		t.Errorf("unexpected success with no plan specified and no serviceclass existing")
	}
	if !strings.Contains(err.Error(), "does not exist, PlanName must be") {
		t.Errorf("did not find expected error")
	}
}

func TestWithNoPlanWorksWithSinglePlan(t *testing.T) {
	sc := newServiceClass("foo", "bar")
	mockSCClient := newMockServiceCatalogClientForTest(sc)
	mockSCClient.AddReactor("get", "serviceclasses", func(action core.Action) (bool, runtime.Object, error) {
		return true, sc, nil
	})
	handler, informerFactory, err := newHandlerForTest(mockSCClient)
	if err != nil {
		t.Errorf("unexpected error initializing handler: %v", err)
	}
	informerFactory.Start(wait.NeverStop)

	instance := newInstance("dummy")
	instance.Spec.ServiceClassName = "foo"
	instance.Spec.PlanName = "bar"

	err = handler.Admit(admission.NewAttributesRecord(&instance, nil, servicecatalog.Kind("Instance").WithVersion("version"), instance.Namespace, instance.Name, servicecatalog.Resource("instances").WithVersion("version"), "", admission.Create, nil))
	if err == nil {
		actions := ""
		for _, action := range mockSCClient.Actions() {
			actions = actions + action.GetVerb() + ":" + action.GetResource().Resource + ":" + action.GetSubresource() + ", "
		}
		t.Errorf("expected error returned from admission handler: %v", actions)
	}
}

/*

// TestAdmissionNamespaceDoesNotExist verifies instance is not admitted if namespace does not exist.
func TestAdmissionNamespaceDoesNotExist(t *testing.T) {
	namespace := "test"
	mockClient := newMockClientForTest()
	mockKubeClient := newMockServiceCatalogClientForTest(map[string]v1.NamespacePhase{})
	mockKubeClient.AddReactor("get", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("nope, out of luck")
	})
	handler, informerFactory, kubeInformerFactory, err := newHandlerForTest(mockClient, mockKubeClient)
	if err != nil {
		t.Errorf("unexpected error initializing handler: %v", err)
	}
	informerFactory.Start(wait.NeverStop)
	kubeInformerFactory.Start(wait.NeverStop)

	instance := newInstance(namespace)
	err = handler.Admit(admission.NewAttributesRecord(&instance, nil, servicecatalog.Kind("Instance").WithVersion("version"), instance.Namespace, instance.Name, servicecatalog.Resource("instances").WithVersion("version"), "", admission.Create, nil))
	if err == nil {
		actions := ""
		for _, action := range mockClient.Actions() {
			actions = actions + action.GetVerb() + ":" + action.GetResource().Resource + ":" + action.GetSubresource() + ", "
		}
		t.Errorf("expected error returned from admission handler: %v", actions)
	}
}

// TestAdmissionNamespaceActive verifies a resource is admitted when the namespace is active.
func TestAdmissionNamespaceActive(t *testing.T) {
	namespace := "test"
	mockClient := newMockClientForTest()
	mockKubeClient := newMockServiceCatalogClientForTest(map[string]v1.NamespacePhase{
		namespace: v1.NamespaceActive,
	})
	handler, informerFactory, kubeInformerFactory, err := newHandlerForTest(mockClient, mockKubeClient)
	if err != nil {
		t.Errorf("unexpected error initializing handler: %v", err)
	}
	informerFactory.Start(wait.NeverStop)
	kubeInformerFactory.Start(wait.NeverStop)

	instance := newInstance(namespace)
	err = handler.Admit(admission.NewAttributesRecord(&instance, nil, servicecatalog.Kind("Instance").WithVersion("version"), instance.Namespace, instance.Name, servicecatalog.Resource("instances").WithVersion("version"), "", admission.Create, nil))
	if err != nil {
		t.Errorf("unexpected error returned from admission handler")
	}
}

// TestAdmissionNamespaceTerminating verifies a resource is not created when the namespace is terminating.
func TestAdmissionNamespaceTerminating(t *testing.T) {
	namespace := "test"
	mockClient := newMockClientForTest()
	mockKubeClient := newMockServiceCatalogClientForTest(map[string]v1.NamespacePhase{
		namespace: v1.NamespaceTerminating,
	})
	handler, informerFactory, kubeInformerFactory, err := newHandlerForTest(mockClient, mockKubeClient)
	if err != nil {
		t.Errorf("unexpected error initializing handler: %v", err)
	}
	informerFactory.Start(wait.NeverStop)
	kubeInformerFactory.Start(wait.NeverStop)

	instance := newInstance(namespace)
	err = handler.Admit(admission.NewAttributesRecord(&instance, nil, servicecatalog.Kind("Instance").WithVersion("version"), instance.Namespace, instance.Name, servicecatalog.Resource("instances").WithVersion("version"), "", admission.Create, nil))
	if err == nil {
		t.Errorf("Expected error rejecting creates in a namespace when it is terminating")
	}

	// verify update operations in the namespace can proceed
	err = handler.Admit(admission.NewAttributesRecord(&instance, nil, servicecatalog.Kind("Instance").WithVersion("version"), instance.Namespace, instance.Name, servicecatalog.Resource("instances").WithVersion("version"), "", admission.Update, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler: %v", err)
	}

	// verify delete operations in the namespace can proceed
	err = handler.Admit(admission.NewAttributesRecord(nil, nil, servicecatalog.Kind("Instance").WithVersion("version"), instance.Namespace, instance.Name, servicecatalog.Resource("instances").WithVersion("version"), "", admission.Delete, nil))
	if err != nil {
		t.Errorf("Unexpected error returned from admission handler: %v", err)
	}
}
*/
