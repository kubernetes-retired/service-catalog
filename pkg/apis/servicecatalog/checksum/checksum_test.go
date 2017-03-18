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

package checksum

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1alpha1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/checksum/unversioned"
	checksumv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/checksum/versioned/v1alpha1"
)

func TestInstanceChecksum(t *testing.T) {
	instanceSpec := servicecatalog.InstanceSpec{
		ServiceClassName: "blorb",
		PlanName:         "plumbus",
		OSBGUID:          "138177saf87)87fs08f7ASfAS*7",
	}

	unversionedChecksum := unversioned.InstanceSpecChecksum(instanceSpec)

	versionedInstanceSpec := v1alpha1.InstanceSpec{}
	v1alpha1.Convert_servicecatalog_InstanceSpec_To_v1alpha1_InstanceSpec(&instanceSpec, &versionedInstanceSpec, nil /* conversionScope */)

	versionedChecksum := checksumv1alpha1.InstanceSpecChecksum(versionedInstanceSpec)

	if e, a := unversionedChecksum, versionedChecksum; e != a {
		t.Fatalf("versioned and unversioned checksums should match; expected %v, got %v", e, a)
	}
}
