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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Broker provides a unifying layer of cluster and namespace scoped broker resources.
type Broker interface {

	// GetName returns the broker's name.
	GetName() string

	// GetNamespace returns the broker's namespace, or "" if it's cluster-scoped.
	GetNamespace() string

	// GetURL returns the broker's URL.
	GetURL() string

	// GetStatus returns the broker's status.
	GetStatus() v1beta1.CommonServiceBrokerStatus
}

// Deregister deletes a broker
func (sdk *SDK) Deregister(brokerName string) error {
	err := sdk.ServiceCatalog().ClusterServiceBrokers().Delete(brokerName, &v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("deregister request failed (%s)", err)
	}

	return nil
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
			if errors.IsNotFound(err) {
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

// Register creates a broker
func (sdk *SDK) Register(brokerName string, url string) (*v1beta1.ClusterServiceBroker, error) {
	request := &v1beta1.ClusterServiceBroker{
		ObjectMeta: v1.ObjectMeta{
			Name: brokerName,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}
	result, err := sdk.ServiceCatalog().ClusterServiceBrokers().Create(request)
	if err != nil {
		return nil, fmt.Errorf("register request failed (%s)", err)
	}

	return result, nil
}

// Sync or relist a broker to refresh its catalog metadata.
func (sdk *SDK) Sync(name string, retries int) error {
	for j := 0; j < retries; j++ {
		catalog, err := sdk.RetrieveBroker(name)
		if err != nil {
			return err
		}

		catalog.Spec.RelistRequests = catalog.Spec.RelistRequests + 1

		_, err = sdk.ServiceCatalog().ClusterServiceBrokers().Update(catalog)
		if err == nil {
			return nil
		}
		if !errors.IsConflict(err) {
			return fmt.Errorf("could not sync service broker (%s)", err)
		}
	}

	return fmt.Errorf("could not sync service broker after %d tries", retries)
}
