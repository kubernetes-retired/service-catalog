/*
Copyright 2018 The Kubernetes Authors.

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

package servicecatalog

import (
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RetrieveSecretByBinding gets the secret associated with a binding
// A nil secret is returned without error when the secret has not been created by Service Catalog yet.
// An error is returned when the binding is Ready but the secret could not be retrieved.
func (sdk *SDK) RetrieveSecretByBinding(binding *v1beta1.ServiceBinding) (*corev1.Secret, error) {
	cond := GetBindingStatusCondition(binding.Status)
	isReady := cond.Type == v1beta1.ServiceBindingConditionReady &&
		cond.Status == v1beta1.ConditionTrue

	secret, err := sdk.Core().Secrets(binding.Namespace).Get(binding.Spec.SecretName, metav1.GetOptions{})
	if err != nil {
		// It's expected to not have the secret until the binding is ready
		if !isReady && errors.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("unable to get secret %s/%s (%s)", binding.Namespace, binding.Spec.SecretName, err)
	}

	return secret, nil
}
