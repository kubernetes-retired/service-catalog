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
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

func appendToErrListAndLog(list field.ErrorList, toAppend ...*field.Error) field.ErrorList {
	if len(toAppend) > 0 {
		logErrList(toAppend...)
		list = append(list, toAppend...)
	}
	return list
}

func logErrList(list ...*field.Error) {
	for i, err := range list {
		glog.Infof("API Validation Error %d: %#v", i, *err)
	}
}
