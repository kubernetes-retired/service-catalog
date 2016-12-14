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

	sbmodel "github.com/kubernetes-incubator/service-catalog/model/service_broker"
	scmodel "github.com/kubernetes-incubator/service-catalog/model/service_controller"
)

// ConvertCredential converts a Credential object from the broker model to the
// controller model.
func ConvertCredential(in *sbmodel.Credential) (*scmodel.Credential, error) {
	j, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	out := &scmodel.Credential{}
	if err := json.Unmarshal([]byte(j), out); err != nil {
		return nil, err
	}

	return out, nil
}

// ConvertCatalog converts a Catalog object from the broker model to the
// controller model.
func ConvertCatalog(in *sbmodel.Catalog) (*scmodel.Catalog, error) {
	j, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	out := &scmodel.Catalog{}
	if err := json.Unmarshal([]byte(j), out); err != nil {
		return nil, err
	}

	return out, nil
}
