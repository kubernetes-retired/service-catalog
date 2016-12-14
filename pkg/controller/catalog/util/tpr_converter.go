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

package util

import (
	"encoding/json"
	"log"

	"k8s.io/client-go/1.5/pkg/runtime"
)

// TPRObjectToSCObject converts a Kubernetes Third Party Resource
// object into a Service Controller model object
func TPRObjectToSCObject(o runtime.Object, object interface{}) error {
	m, err := json.Marshal(o)
	if err != nil {
		log.Printf("Failed to marshal %#v : %v", o, err)
		return err
	}

	err = json.Unmarshal(m, &object)
	if err != nil {
		log.Printf("Failed to unmarshal: %v\n", err)
		return err
	}
	return nil
}

// SCObjectToTPRObject converts a Service Controller model object into
// a Kubernetes Third Party Resource object
func SCObjectToTPRObject(object interface{}) (*runtime.Unstructured, error) {
	m, err := json.Marshal(object)
	if err != nil {
		log.Printf("Failed to marshal %#v : %v", object, err)
		return nil, err
	}
	var ret runtime.Unstructured
	err = json.Unmarshal(m, &ret)
	if err != nil {
		log.Printf("Failed to unmarshal: %v\n", err)
		return nil, err
	}
	return &ret, nil
}
