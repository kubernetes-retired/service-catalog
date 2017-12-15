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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

func TestValidateClusterID(t *testing.T) {
	cases := []struct {
		name  string
		id    *sc.ClusterID
		valid bool
	}{
		{
			"happy path, name is cluster-id, ID is set",
			&sc.ClusterID{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-id",
				},
				ID: "some-id",
			},
			true,
		},
		{
			"missing ID",
			&sc.ClusterID{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-id",
				},
			},
			false,
		},
		{
			"name not cluster-id",
			&sc.ClusterID{
				ObjectMeta: metav1.ObjectMeta{
					Name: "steve",
				},
			},
			false,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			errs := ValidateClusterID(tc.id)
			if len(errs) != 0 && tc.valid {
				t.Errorf("%v: unexpected error: %v", tc.name, errs)
			} else if len(errs) == 0 && !tc.valid {
				t.Errorf("%v: unexpected success", tc.name)
			}
		})
	}
}
