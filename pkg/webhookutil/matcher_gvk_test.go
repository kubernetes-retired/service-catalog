/*
Copyright 2019 The Kubernetes Authors.

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

package webhookutil_test

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1beta1 "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
)

func TestMatchKinds(t *testing.T) {
	t.Run("Should return no error for same GVK", func(t *testing.T) {
		// given
		reqGVK := metav1.GroupVersionKind{
			Kind:    "Deployment",
			Group:   "apps",
			Version: "v1beta1",
		}

		deployObject := &corev1beta1.Deployment{}

		// when
		err := webhookutil.MatchKinds(deployObject, reqGVK)

		// then
		assert.NoError(t, err)
	})

	t.Run("Should return error for different GVK", func(t *testing.T) {
		// given
		reqGVK := metav1.GroupVersionKind{
			Kind:    "Pod",
			Group:   "apps",
			Version: "v1beta1",
		}

		deployObject := &corev1beta1.Deployment{}

		// when
		err := webhookutil.MatchKinds(deployObject, reqGVK)

		// then
		assert.EqualError(t, err, "type mismatch: want: apps/v1beta1, Kind=Deployment got: apps/v1beta1, Kind=Pod")
	})

	t.Run("Should return error if GVK is not registered", func(t *testing.T) {
		// given
		reqGVK := metav1.GroupVersionKind{}
		csbObject := &v1beta1.ClusterServiceBroker{}

		// when
		err := webhookutil.MatchKinds(csbObject, reqGVK)

		// then
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "no kind is registered for the type v1beta1.ClusterServiceBroker in scheme"))
	})
}
