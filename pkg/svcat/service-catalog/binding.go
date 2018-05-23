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
	"math"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// RetrieveBindings lists all bindings in a namespace.
func (sdk *SDK) RetrieveBindings(ns string) (*v1beta1.ServiceBindingList, error) {
	bindings, err := sdk.ServiceCatalog().ServiceBindings(ns).List(v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list bindings in %s", ns)
	}

	return bindings, nil
}

// RetrieveBinding gets a binding by its name.
func (sdk *SDK) RetrieveBinding(ns, name string) (*v1beta1.ServiceBinding, error) {
	binding, err := sdk.ServiceCatalog().ServiceBindings(ns).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get binding '%s.%s'", ns, name)
	}
	return binding, nil
}

// RetrieveBindingsByInstance gets all child bindings for an instance.
func (sdk *SDK) RetrieveBindingsByInstance(instance *v1beta1.ServiceInstance,
) ([]v1beta1.ServiceBinding, error) {
	// Not using a filtered list operation because it's not supported yet.
	results, err := sdk.ServiceCatalog().ServiceBindings(instance.Namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to search bindings")
	}

	var bindings []v1beta1.ServiceBinding
	for _, binding := range results.Items {
		if binding.Spec.ServiceInstanceRef.Name == instance.Name {
			bindings = append(bindings, binding)
		}
	}

	return bindings, nil
}

// Bind an instance to a secret.
func (sdk *SDK) Bind(namespace, bindingName, externalID, instanceName, secretName string,
	params interface{}, secrets map[string]string) (*v1beta1.ServiceBinding, error) {

	// Manually defaulting the name of the binding
	// I'm not doing the same for the secret since the API handles defaulting that value.
	if bindingName == "" {
		bindingName = instanceName
	}

	request := &v1beta1.ServiceBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceBindingSpec{
			ExternalID: externalID,
			ServiceInstanceRef: v1beta1.LocalObjectReference{
				Name: instanceName,
			},
			SecretName:     secretName,
			Parameters:     BuildParameters(params),
			ParametersFrom: BuildParametersFrom(secrets),
		},
	}

	result, err := sdk.ServiceCatalog().ServiceBindings(namespace).Create(request)
	if err != nil {
		return nil, errors.Wrap(err, "bind request failed")
	}

	return result, nil
}

// Unbind deletes all bindings associated to an instance.
func (sdk *SDK) Unbind(ns, instanceName string) ([]v1beta1.ServiceBinding, error) {
	instance, err := sdk.RetrieveInstance(ns, instanceName)
	if err != nil {
		return nil, err
	}
	bindings, err := sdk.RetrieveBindingsByInstance(instance)
	if err != nil {
		return nil, err
	}
	var g sync.WaitGroup
	errs := make(chan error, len(bindings))
	deletedBindings := make(chan v1beta1.ServiceBinding, len(bindings))
	for _, binding := range bindings {
		g.Add(1)
		go func(binding v1beta1.ServiceBinding) {
			defer g.Done()
			err := sdk.DeleteBinding(binding.Namespace, binding.Name)
			if err == nil {
				deletedBindings <- binding
			}
			errs <- err
		}(binding)
	}

	g.Wait()
	close(errs)
	close(deletedBindings)

	// Collect any errors that occurred into a single formatted error
	bindErr := &multierror.Error{
		ErrorFormat: func(errors []error) string {
			return joinErrors("could not remove some bindings:", errors, "\n  ")
		},
	}
	for err := range errs {
		bindErr = multierror.Append(bindErr, err)
	}

	//Range over the deleted bindings to build a slice to return
	deleted := []v1beta1.ServiceBinding(nil)
	for b := range deletedBindings {
		deleted = append(deleted, b)
	}
	return deleted, bindErr.ErrorOrNil()
}

// DeleteBinding by name.
func (sdk *SDK) DeleteBinding(ns, bindingName string) error {
	err := sdk.ServiceCatalog().ServiceBindings(ns).Delete(bindingName, &v1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "remove binding %s/%s failed", ns, bindingName)
	}
	return nil
}

func joinErrors(groupMsg string, errors []error, sep string, a ...interface{}) string {
	if len(errors) == 0 {
		return ""
	}

	msgs := make([]string, 0, len(errors)+1)
	msgs = append(msgs, fmt.Sprintf(groupMsg, a...))
	for _, err := range errors {
		msgs = append(msgs, err.Error())
	}

	return strings.Join(msgs, sep)
}

// BindingParentHierarchy retrieves all ancestor resources of a binding.
func (sdk *SDK) BindingParentHierarchy(binding *v1beta1.ServiceBinding,
) (*v1beta1.ServiceInstance, *v1beta1.ClusterServiceClass, *v1beta1.ClusterServicePlan, *v1beta1.ClusterServiceBroker, error) {
	instance, err := sdk.RetrieveInstanceByBinding(binding)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	class, plan, err := sdk.InstanceToServiceClassAndPlan(instance)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	broker, err := sdk.RetrieveBrokerByClass(class)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return instance, class, plan, broker, nil
}

// GetBindingStatusCondition returns the last condition on a binding status.
// When no conditions exist, an empty condition is returned.
func GetBindingStatusCondition(status v1beta1.ServiceBindingStatus) v1beta1.ServiceBindingCondition {
	if len(status.Conditions) > 0 {
		return status.Conditions[len(status.Conditions)-1]
	}
	return v1beta1.ServiceBindingCondition{}
}

// WaitForBinding waits for the instance to complete the current operation (or fail).
func (sdk *SDK) WaitForBinding(ns, name string, interval time.Duration, timeout *time.Duration) (binding *v1beta1.ServiceBinding, err error) {
	if timeout == nil {
		notimeout := time.Duration(math.MaxInt64)
		timeout = &notimeout
	}

	err = wait.PollImmediate(interval, *timeout,
		func() (bool, error) {
			binding, err = sdk.RetrieveBinding(ns, name)
			if nil != err {
				if apierrors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}

			if len(binding.Status.Conditions) == 0 {
				return false, nil
			}

			isDone := (sdk.IsBindingReady(binding) || sdk.IsBindingFailed(binding)) && !binding.Status.AsyncOpInProgress
			return isDone, nil
		},
	)

	return binding, err
}

// IsBindingReady returns if the instance is in the Ready status.
func (sdk *SDK) IsBindingReady(binding *v1beta1.ServiceBinding) bool {
	return sdk.BindingHasStatus(binding, v1beta1.ServiceBindingConditionReady)
}

// IsBindingFailed returns if the instance is in the Failed status.
func (sdk *SDK) IsBindingFailed(binding *v1beta1.ServiceBinding) bool {
	return sdk.BindingHasStatus(binding, v1beta1.ServiceBindingConditionFailed)
}

// BindingHasStatus returns if the instance is in the specified status.
func (sdk *SDK) BindingHasStatus(binding *v1beta1.ServiceBinding, status v1beta1.ServiceBindingConditionType) bool {
	for _, cond := range binding.Status.Conditions {
		if cond.Type == status &&
			cond.Status == v1beta1.ConditionTrue {
			return true
		}
	}

	return false
}
