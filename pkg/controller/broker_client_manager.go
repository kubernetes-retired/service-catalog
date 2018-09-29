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

package controller

import (
	"fmt"
	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"reflect"
	"sync"
)

// BrokerKey defines a key which points to a broker (cluster wide or namespaced)
type BrokerKey struct {
	name      string
	namespace string
}

// IsClusterScoped whether this broker key points to cluster scoped service broker.
func (bk *BrokerKey) IsClusterScoped() bool {
	return bk.namespace == ""
}

// String returns string representation of the broker key
func (bk *BrokerKey) String() string {
	if bk.IsClusterScoped() {
		return bk.name
	}
	return fmt.Sprintf("%s/%s", bk.namespace, bk.name)
}

// NewServiceBrokerKey creates a BrokerKey instance which points to namespaced broker
func NewServiceBrokerKey(namespace, name string) BrokerKey {
	return BrokerKey{
		namespace: namespace,
		name:      name,
	}
}

// NewClusterServiceBrokerKey creates a BrokerKey instance which points to cluster wide broker
func NewClusterServiceBrokerKey(name string) BrokerKey {
	return BrokerKey{
		namespace: "",
		name:      name,
	}
}

// BrokerClientManager stores OSB client instances per broker
type BrokerClientManager struct {
	mu      sync.RWMutex
	clients map[BrokerKey]clientWithConfig

	brokerClientCreateFunc osb.CreateFunc
}

// NewBrokerClientManager creates BrokerClientManager instance
func NewBrokerClientManager(brokerClientCreateFunc osb.CreateFunc) *BrokerClientManager {
	return &BrokerClientManager{
		clients:                map[BrokerKey]clientWithConfig{},
		brokerClientCreateFunc: brokerClientCreateFunc,
	}
}

// UpdateBrokerClient creates new broker client if necessary (the ClientConfig has changed or there is no client for the broker),
// the method returns created or stored osb.Client instance.
func (m *BrokerClientManager) UpdateBrokerClient(brokerKey BrokerKey, clientConfig *osb.ClientConfiguration) (osb.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, found := m.clients[brokerKey]

	if !found || configHasChanged(existing.clientConfig, clientConfig) {
		glog.V(4).Infof("Updating OSB client for broker %s, URL: %s", brokerKey.String(), clientConfig.URL)
		return m.createClient(brokerKey, clientConfig)
	}

	return existing.OSBClient, nil
}

// RemoveBrokerClient removes broker client broker
func (m *BrokerClientManager) RemoveBrokerClient(brokerKey BrokerKey) {
	m.mu.Lock()
	defer m.mu.Unlock()

	glog.V(4).Info("Removing OSB client for broker %s", brokerKey.String())
	delete(m.clients, brokerKey)
}

// BrokerClient returns broker client for a broker specified by the brokerKey
func (m *BrokerClientManager) BrokerClient(brokerKey BrokerKey) (osb.Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	existing, found := m.clients[brokerKey]
	return existing.OSBClient, found
}

func (m *BrokerClientManager) createClient(brokerKey BrokerKey, clientConfig *osb.ClientConfiguration) (osb.Client, error) {
	client, err := m.brokerClientCreateFunc(clientConfig)
	if err != nil {
		return nil, err
	}

	m.clients[brokerKey] = clientWithConfig{
		OSBClient:    client,
		clientConfig: clientConfig,
	}
	return client, nil
}

func configHasChanged(cfg1 *osb.ClientConfiguration, cfg2 *osb.ClientConfiguration) bool {
	return !reflect.DeepEqual(cfg1, cfg2)
}

type clientWithConfig struct {
	OSBClient    osb.Client
	clientConfig *osb.ClientConfiguration
}
