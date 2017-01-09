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

package storage

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/watch"
	// Need this for gcp auth
	_ "k8s.io/client-go/1.5/kubernetes"
)

type tprStorage struct {
	watcher *watch.Watcher
}

// CreateTPRStorage creates an instance of Storage backed by Kubernetes
// third-party resources.
func CreateTPRStorage(w *watch.Watcher) Storage {
	return &tprStorage{
		watcher: w,
	}
}

func (t *tprStorage) Brokers() BrokerStorage {
	return newTPRStorageBroker(t.watcher)
}

func (t *tprStorage) ServiceClasses() ServiceClassStorage {
	return newTPRStorageServiceClass(t.watcher)
}

func (t *tprStorage) Instances(ns string) InstanceStorage {
	return newTPRStorageInstance(t.watcher, ns)
}

func (t *tprStorage) Bindings(ns string) BindingStorage {
	return newTPRStorageBinding(t.watcher, ns)
}
