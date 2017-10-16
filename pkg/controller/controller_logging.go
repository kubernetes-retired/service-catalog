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

package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1beta1"
	listers "github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"k8s.io/gengo/testdata/a/b"
)

// Logging should always be in the form:
// <source>: <msg>
//
// <source> is in the following form: <type> "<Namespace>/<Name>": <msg>
//
// <msg> might have a type that needs K8S or External context, that form is:
// <type> (K8S: <K8S-Type-Name> ExternalName: <External-Type-Name>)

func (c *controller) logContext(name string) clientset.Interface {
	client, err := b.Client(name)
	if err != nil {
		glog.Fatal(err)
	}
	return client
}
