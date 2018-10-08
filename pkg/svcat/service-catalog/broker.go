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
	"io/ioutil"
	"math"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Broker provides a unifying layer of cluster and namespace scoped broker resources.
type Broker interface {

	// GetName returns the broker's name.
	GetName() string

	// GetNamespace returns the broker's namespace, or "" if it's cluster-scoped.
	GetNamespace() string

	// GetURL returns the broker's URL.
	GetURL() string

	// GetSpec returns the broker's spec.
	GetSpec() v1beta1.CommonServiceBrokerSpec

	// GetStatus returns the broker's status.
	GetStatus() v1beta1.CommonServiceBrokerStatus
}

// Deregister deletes a broker
func (sdk *SDK) Deregister(brokerName string, scopeOpts *ScopeOptions) error {
	if scopeOpts.Scope.Matches(NamespaceScope) {
		err := sdk.ServiceCatalog().ServiceBrokers(scopeOpts.Namespace).Delete(brokerName, &v1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("deregister request failed (%s)", err)
		}
		return nil
	} else if scopeOpts.Scope.Matches(ClusterScope) {
		err := sdk.ServiceCatalog().ClusterServiceBrokers().Delete(brokerName, &v1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("deregister request failed (%s)", err)
		}
		return nil
	}
	return fmt.Errorf("cannot deregister broker, unrecognized scope provided (%s)", scopeOpts.Scope)
}

// RetrieveBrokers lists all brokers defined in the cluster.
func (sdk *SDK) RetrieveBrokers(opts ScopeOptions) ([]Broker, error) {
	var brokers []Broker

	if opts.Scope.Matches(ClusterScope) {
		csb, err := sdk.ServiceCatalog().ClusterServiceBrokers().List(v1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to list cluster-scoped brokers (%s)", err)
		}
		for _, b := range csb.Items {
			broker := b
			brokers = append(brokers, &broker)
		}
	}

	if opts.Scope.Matches(NamespaceScope) {
		sb, err := sdk.ServiceCatalog().ServiceBrokers(opts.Namespace).List(v1.ListOptions{})
		if err != nil {
			// Gracefully handle when the feature-flag for namespaced broker resources isn't enabled on the server.
			if apierrors.IsNotFound(err) {
				return brokers, nil
			}
			return nil, fmt.Errorf("unable to list brokers in %q (%s)", opts.Namespace, err)
		}
		for _, b := range sb.Items {
			broker := b
			brokers = append(brokers, &broker)
		}
	}

	return brokers, nil
}

// RetrieveBroker gets a broker by its name.
func (sdk *SDK) RetrieveBroker(name string) (*v1beta1.ClusterServiceBroker, error) {
	broker, err := sdk.ServiceCatalog().ClusterServiceBrokers().Get(name, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get broker '%s'", name)
	}

	return broker, nil
}

// RetrieveNamespacedBroker gets a broker by its name & namespace.
func (sdk *SDK) RetrieveNamespacedBroker(namespace string, name string) (*v1beta1.ServiceBroker, error) {
	broker, err := sdk.ServiceCatalog().ServiceBrokers(namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get broker '%s' (%s)", name, err)
	}

	return broker, nil
}

// RetrieveBrokerByClass gets the parent broker of a class.
func (sdk *SDK) RetrieveBrokerByClass(class *v1beta1.ClusterServiceClass,
) (*v1beta1.ClusterServiceBroker, error) {
	brokerName := class.Spec.ClusterServiceBrokerName
	broker, err := sdk.ServiceCatalog().ClusterServiceBrokers().Get(brokerName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return broker, nil
}

//Register creates a broker
func (sdk *SDK) Register(brokerName string, url string, opts *RegisterOptions, scopeOpts *ScopeOptions) (Broker, error) {
	var err error
	var caBytes []byte
	if opts.CAFile != "" {
		caBytes, err = ioutil.ReadFile(opts.CAFile)
		if err != nil {
			return nil, fmt.Errorf("Error opening CA file: %v", err.Error())
		}

	}
	objectMeta := v1.ObjectMeta{Name: brokerName}
	commonServiceBrokerSpec := v1beta1.CommonServiceBrokerSpec{
		CABundle:              caBytes,
		InsecureSkipTLSVerify: opts.SkipTLS,
		RelistBehavior:        opts.RelistBehavior,
		RelistDuration:        opts.RelistDuration,
		URL:                   url,
		CatalogRestrictions: &v1beta1.CatalogRestrictions{
			ServiceClass: opts.ClassRestrictions,
			ServicePlan:  opts.PlanRestrictions,
		},
	}
	if scopeOpts.Scope.Matches(ClusterScope) {
		request := &v1beta1.ClusterServiceBroker{
			ObjectMeta: objectMeta,
			Spec: v1beta1.ClusterServiceBrokerSpec{
				CommonServiceBrokerSpec: commonServiceBrokerSpec,
			},
		}
		request.Spec.AuthInfo = &v1beta1.ClusterServiceBrokerAuthInfo{}
		if opts.BasicSecret != "" {
			request.Spec.AuthInfo.Basic = &v1beta1.ClusterBasicAuthConfig{
				SecretRef: &v1beta1.ObjectReference{
					Name:      opts.BasicSecret,
					Namespace: opts.Namespace,
				},
			}
		} else if opts.BearerSecret != "" {
			request.Spec.AuthInfo.Bearer = &v1beta1.ClusterBearerTokenAuthConfig{
				SecretRef: &v1beta1.ObjectReference{
					Name:      opts.BearerSecret,
					Namespace: opts.Namespace,
				},
			}
		}

		result, err := sdk.ServiceCatalog().ClusterServiceBrokers().Create(request)
		if err != nil {
			return nil, fmt.Errorf("register request failed (%s)", err)
		}

		return result, nil
	} //else matches NamespaceScope
	request := &v1beta1.ServiceBroker{
		ObjectMeta: objectMeta,
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: commonServiceBrokerSpec,
		},
	}
	if opts.BasicSecret != "" {
		request.Spec.AuthInfo = &v1beta1.ServiceBrokerAuthInfo{
			Basic: &v1beta1.BasicAuthConfig{
				SecretRef: &v1beta1.LocalObjectReference{
					Name: opts.BasicSecret,
				},
			},
		}
	} else if opts.BearerSecret != "" {
		request.Spec.AuthInfo = &v1beta1.ServiceBrokerAuthInfo{
			Bearer: &v1beta1.BearerTokenAuthConfig{
				SecretRef: &v1beta1.LocalObjectReference{
					Name: opts.BearerSecret,
				},
			},
		}
	}

	result, err := sdk.ServiceCatalog().ServiceBrokers(scopeOpts.Namespace).Create(request)
	if err != nil {
		return nil, fmt.Errorf("register request failed (%s)", err)
	}
	return result, nil
}

// Sync or relist a broker to refresh its broker metadata.
func (sdk *SDK) Sync(name string, scopeOpts ScopeOptions, retries int) error {
	success := false
	var err error

	for j := 0; j < retries && !success; j++ {

		if scopeOpts.Scope.Matches(NamespaceScope) {
			var broker *v1beta1.ServiceBroker
			namespace := scopeOpts.Namespace
			broker, err = sdk.RetrieveNamespacedBroker(namespace, name)
			if err == nil {
				broker.Spec.RelistRequests = broker.Spec.RelistRequests + 1

				_, err = sdk.ServiceCatalog().ServiceBrokers(namespace).Update(broker)
				if err == nil {
					success = true
				}
				if err != nil && !apierrors.IsConflict(err) {
					return fmt.Errorf("could not sync service broker (%s)", err)
				}
			}
		}

		if scopeOpts.Scope.Matches(ClusterScope) {
			var broker *v1beta1.ClusterServiceBroker
			broker, err = sdk.RetrieveBroker(name)
			if err == nil {
				broker.Spec.RelistRequests = broker.Spec.RelistRequests + 1

				_, err = sdk.ServiceCatalog().ClusterServiceBrokers().Update(broker)
				if err == nil {
					success = true
				}
				if err != nil && !apierrors.IsConflict(err) {
					return fmt.Errorf("could not sync service broker (%s)", err)
				}
			}
		}
		// success to update, no need to retry again
		if success {
			break
		}
	}

	if !success {
		return fmt.Errorf("could not sync service broker %s (%s)", name, err)
	}

	return nil
}

// WaitForBroker waits for the specified broker to be Ready or Failed
func (sdk *SDK) WaitForBroker(name string, interval time.Duration, timeout *time.Duration) (broker Broker, err error) {
	if timeout == nil {
		notimeout := time.Duration(math.MaxInt64)
		timeout = &notimeout
	}
	err = wait.PollImmediate(interval, *timeout,
		func() (bool, error) {
			broker, err = sdk.RetrieveBroker(name)
			if err != nil {
				if apierrors.IsNotFound(errors.Cause(err)) {
					err = nil
				}
				return false, err
			}

			isDone := sdk.IsBrokerReady(broker) || sdk.IsBrokerFailed(broker)
			return isDone, nil
		})
	return broker, err
}

// IsBrokerReady returns if the broker is in the Ready status.
func (sdk *SDK) IsBrokerReady(broker Broker) bool {
	return sdk.BrokerHasStatus(broker, v1beta1.ServiceBrokerConditionReady)
}

// IsBrokerFailed returns if the broker is in the Failed status.
func (sdk *SDK) IsBrokerFailed(broker Broker) bool {
	return sdk.BrokerHasStatus(broker, v1beta1.ServiceBrokerConditionFailed)
}

// BrokerHasStatus returns if the broker is in the specified status.
func (sdk *SDK) BrokerHasStatus(broker Broker, status v1beta1.ServiceBrokerConditionType) bool {
	for _, cond := range broker.GetStatus().Conditions {
		if cond.Type == status &&
			cond.Status == v1beta1.ConditionTrue {
			return true
		}
	}

	return false
}
