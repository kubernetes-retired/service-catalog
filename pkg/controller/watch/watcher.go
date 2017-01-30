/*
Copyright 2016 The Kubernetes Authors.

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

package watch

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/1.5/dynamic"
	// Need this for gcp auth
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/watch"
)

type resourceType int

// These define the Third Party Resources that we can handle or operate on.
const (
	ServiceInstance = iota
	ServiceBinding
	ServiceBroker
	ServiceClass
)

func (rt *resourceType) name() string {
	return resourceTypeNames[*rt]
}

var resourceTypeNames = []string{
	"ServiceInstance",
	"ServiceBinding",
	"ServiceBroker",
	"ServiceClass",
}

// These resources _must_ exist in the cluster before proceeding.
var resourceTypes = []resourceType{ServiceInstance, ServiceBinding, ServiceBroker, ServiceClass}

const (
	// GroupName is a name of a Kubernetes API extension implemented by the service catalog.
	GroupName = "catalog.k8s.io"

	// APIVersion is a version of the Kubernetes API extension implemented by the service catalog.
	APIVersion = "v1alpha1"

	// FullAPIVersion is a fully qualified name of the Kubernetes API extension implemented by the service catalog.
	FullAPIVersion = GroupName + "/" + APIVersion

	// ServiceBrokerKind is a name of a Service Broker resource, a Kubernetes third party resource.
	ServiceBrokerKind = "ServiceBroker"

	// ServiceBindingKind is a name of a Service Binding resource, a Kubernetes third party resource.
	ServiceBindingKind = "ServiceBinding"

	// ServiceClassKind is a name of a Service Class resource, a Kubernetes third party resource.
	ServiceClassKind = "ServiceClass"

	// ServiceInstanceKind is a name of a Service Instance resource, a Kubernetes third party resource.
	ServiceInstanceKind = "ServiceInstance"
)

var thirdPartyResourceTypes = map[string]v1beta1.ThirdPartyResource{
	"service-broker.catalog.k8s.io": {
		ObjectMeta:  v1.ObjectMeta{Name: "service-broker.catalog.k8s.io"},
		Description: "A Service Broker representation. Adds a service broker and fetches its catalog",
		Versions:    []v1beta1.APIVersion{{Name: "v1alpha1"}},
	},
	"service-class.catalog.k8s.io": {
		ObjectMeta:  v1.ObjectMeta{Name: "service-class.catalog.k8s.io"},
		Description: "A Service Type representation. Something that a customer can instantiate",
		Versions:    []v1beta1.APIVersion{{Name: "v1alpha1"}},
	},
	"service-instance.catalog.k8s.io": {
		ObjectMeta:  v1.ObjectMeta{Name: "service-instance.catalog.k8s.io"},
		Description: "A Service Instance representation, creates a Service Instance",
		Versions:    []v1beta1.APIVersion{{Name: "v1alpha1"}},
	},
	"service-binding.catalog.k8s.io": {
		ObjectMeta:  v1.ObjectMeta{Name: "service-binding.catalog.k8s.io"},
		Description: "A Service Binding representation, creates a Service Binding",
		Versions:    []v1beta1.APIVersion{{Name: "v1alpha1"}},
	},
}

var serviceInstanceResource = unversioned.APIResource{
	Name:       "serviceinstances",
	Kind:       ServiceInstanceKind,
	Namespaced: true,
}

var serviceBindingResource = unversioned.APIResource{
	Name:       "servicebindings",
	Kind:       ServiceBindingKind,
	Namespaced: true,
}

var serviceBrokerResource = unversioned.APIResource{
	Name:       "servicebrokers",
	Kind:       ServiceBrokerKind,
	Namespaced: true,
}

var serviceClassResource = unversioned.APIResource{
	Name:       "serviceclasses",
	Kind:       ServiceClassKind,
	Namespaced: true,
}

type watchCallback func(watch.Event) error

// Watcher watches for Kubernetes events, such as creation of resources, and
// performs custom operations.
type Watcher struct {
	dynClient *dynamic.Client
}

// NewWatcher creates a new Watcher. kubeconfig specifies the kubeconfig file to
// use, if config = "" uses incluster config.
func NewWatcher(k8sClient *kubernetes.Clientset, dynClient *dynamic.Client) (*Watcher, error) {
	err := initCluster(k8sClient)
	if err != nil {
		glog.Errorf("Failed to initialize cluster: %v\n", err)
		return nil, err
	}

	err = checkCluster(dynClient)
	if err != nil {
		glog.Errorf("Cluster does not seem to in correct working order: %v\n", err)
	}

	return &Watcher{
		dynClient: dynClient,
	}, nil
}

func initCluster(clientset *kubernetes.Clientset) error {
	for k, v := range thirdPartyResourceTypes {
		glog.Infof("Checking for existence of %s\n", k)
		_, err := clientset.Extensions().ThirdPartyResources().Get(k)
		if err == nil {
			glog.Errorf("Found existing TPR %s\n", k)
			continue
		}

		glog.Infof("Creating Third Party Resource Type: %s\n", k)
		_, err = clientset.Extensions().ThirdPartyResources().Create(&v)
		if err != nil {
			glog.Errorf("Failed to create Third Party Resource Type: %s : %v\n", k, err)
			return err
		}
		glog.Infof("Created TPR: %s\n", k)
		// There can be a delay, so poll until it's ready to go...
		for i := 0; i < 30; i++ {
			_, err = clientset.Extensions().ThirdPartyResources().Get(k)
			if err == nil {
				glog.Infof("TPR is ready %s\n", k)
				break
			}
			glog.Infof("TPR: %s is not ready yet... waiting...\n", k)
			time.Sleep(1 * time.Second)
		}
	}

	thirdparty, err := clientset.Extensions().ThirdPartyResources().List(api.ListOptions{})
	if err != nil {
		return err
	}
	for _, apis := range thirdparty.Items {
		glog.Infof("Thirdparty: %+v\n", apis)
	}
	return nil
}

func checkCluster(client *dynamic.Client) error {
	glog.Infoln("initCluster")

	for _, rt := range resourceTypes {
		c := getResourceClient(client, rt, "default")
		if c == nil {
			return fmt.Errorf("Failed to get a client %d", rt)
		}

		glog.Infof("Checking resource type %s for readiness for listing\n", rt.name())
		ok := false
		for i := 0; i < 30; i++ {
			_, err := c.List(&v1.ListOptions{})
			if err == nil {
				glog.Infof("Successful list for %s, continuing\n", rt.name())
				ok = true
				break
			}
			glog.Errorf("Failed to list for %s... waiting...\n", rt.name())
			time.Sleep(1 * time.Second)
		}
		if !ok {
			glog.Errorf("Can't list %s, bailing...\n", rt.name())
			return fmt.Errorf("Third Party Resource Type %s is not ready", rt.name())
		}
	}
	return nil
}

// Watch starts a watch for ResourceType t and on events will call the wcb.
func (w *Watcher) Watch(t resourceType, ns string, wcb watchCallback) {
	rc := w.GetResourceClient(t, ns)
	go thirdPartyWatcher(newRealDynamicResourceClient(rc), wcb)
}

func thirdPartyWatcher(rc dynamicResourceClient, wcb watchCallback) {
	for {
		// First do List on the resource to bring things up to date.
		/*
			l, err := rc.List(&v1.ListOptions{})
			for _, o := range l.Items {
				glog.Infof("Found Third Party Resource name: %s\n", o.Name)
				event := watch.Event{
					Type:   watch.Added,
					Object: &o,
				}
				wcb(event)
			}

		*/
		w, err := rc.Watch(&v1.ListOptions{})
		if err != nil {
			glog.Errorf("Failed to start a watch: %v\n", err)
			continue
		}
		c := w.ResultChan()

		glog.Infoln("Entering watch loop")
		done := false
		for {
			select {
			case <-time.After(1 * time.Minute):
				glog.Infoln("*** select heartbeat ***")
			case e := <-c:
				glog.Infof("Watch called with event Type: %s\n", e.Type)
				if e.Object == nil {
					glog.Infoln("Watch appears to have failed, restarting watch loop...")
					done = true
				} else {
					wcb(e)
				}
			}
			if done {
				glog.Infoln("Bailing from select for loop")
				break
			}
		}
	}
}

// GetResourceClient returns a dynamic resource client for interacting with the given resource type.
func (w *Watcher) GetResourceClient(t resourceType, namespace string) *dynamic.ResourceClient {
	return getResourceClient(w.dynClient, t, namespace)
}

func getResourceClient(client *dynamic.Client, t resourceType, namespace string) *dynamic.ResourceClient {
	switch t {
	case ServiceInstance:
		return client.Resource(&serviceInstanceResource, namespace)
	case ServiceBinding:
		return client.Resource(&serviceBindingResource, namespace)
	case ServiceBroker:
		return client.Resource(&serviceBrokerResource, namespace)
	case ServiceClass:
		return client.Resource(&serviceClassResource, namespace)
	}
	return nil
}
