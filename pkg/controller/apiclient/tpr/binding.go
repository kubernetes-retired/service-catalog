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

package tpr

import (
	"errors"
	"log"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/util"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/watch"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/runtime"
)

type bindingClient struct {
	watcher *watch.Watcher
	ns      string
}

func newBindingClient(watcher *watch.Watcher, ns string) *bindingClient {
	return &bindingClient{watcher: watcher, ns: ns}
}

func (c *bindingClient) Update(in *servicecatalog.Binding) (*servicecatalog.Binding, error) {
	in.Kind = watch.ServiceBindingKind
	in.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(in)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", in, err)
		return nil, err
	}
	tprObj.SetName(in.Name)
	log.Printf("Updating Binding %s in k8s:\n%v\n", in.Name, tprObj)
	_, err = c.watcher.GetResourceClient(watch.ServiceBinding, "default").Update(tprObj)
	// krancour: Ideally the binding we return is a translation of the updated 3pr
	// as read back from k8s. It doesn't seem worth going through the trouble
	// right now since 3pr storage will be removed eventually. This will at least
	// work well enough in the meantime.
	return in, err
}

// List returns all the bindings
func (c *bindingClient) List() ([]*servicecatalog.Binding, error) {
	l, err := c.watcher.GetResourceClient(watch.ServiceBinding, c.ns).List(&v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var ret []*servicecatalog.Binding
	for _, i := range l.(*runtime.UnstructuredList).Items {
		var tmp servicecatalog.Binding
		err := util.TPRObjectToSCObject(i, &tmp)
		if err != nil {
			log.Printf("Failed to convert object: %v\n", err)
			return nil, err
		}
		ret = append(ret, &tmp)
	}
	return ret, nil
}

func (*bindingClient) Get(string) (*servicecatalog.Binding, error) {
	return nil, errors.New("Not implemented yet")
}

func (c *bindingClient) Create(in *servicecatalog.Binding) (*servicecatalog.Binding, error) {
	in.Kind = watch.ServiceBindingKind
	in.APIVersion = watch.FullAPIVersion
	tprObj, err := util.SCObjectToTPRObject(in)
	if err != nil {
		log.Printf("Failed to convert object %#v : %v", in, err)
		return nil, err
	}
	tprObj.SetName(in.Name)
	log.Printf("Creating binding %s:\n%v\n", in.Name, tprObj)
	_, err = c.watcher.GetResourceClient(watch.ServiceBinding, c.ns).Create(tprObj)
	// krancour: Ideally the binding we return is a translation of the updated 3pr
	// as read back from k8s. It doesn't seem worth going through the trouble
	// right now since 3pr storage will be removed eventually. This will at least
	// work well enough in the meantime.
	return in, err
}

func (*bindingClient) Delete(string) error {
	return errors.New("Not implemented yet")
}
