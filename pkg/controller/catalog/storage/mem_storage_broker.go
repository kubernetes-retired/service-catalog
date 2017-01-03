package storage

import (
	"errors"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
)

var (
	errBrokerAlreadyExists = errors.New("broker already exists")
	errNoSuchBroker        = errors.New("no such broker")
)

type memStorageBroker struct {
	brokers map[string]*servicecatalog.Broker
}

func newMemStorageBroker() *memStorageBroker {
	return &memStorageBroker{brokers: make(map[string]*servicecatalog.Broker)}
}

func (m *memStorageBroker) List() ([]*servicecatalog.Broker, error) {
	ret := make([]*servicecatalog.Broker, len(m.brokers))
	i := 0
	for _, val := range m.brokers {
		ret[i] = val
		i++
	}
	return ret, nil
}

func (m *memStorageBroker) Get(name string) (*servicecatalog.Broker, error) {
	ret, ok := m.brokers[name]
	if !ok {
		return nil, errNoSuchBroker
	}
	return ret, nil
}

func (m *memStorageBroker) Create(br *servicecatalog.Broker) (*servicecatalog.Broker, error) {
	if _, err := m.Get(br.Name); err == nil {
		return nil, errBrokerAlreadyExists
	}
	m.brokers[br.Name] = br
	return br, nil
}

func (m *memStorageBroker) Update(br *servicecatalog.Broker) (*servicecatalog.Broker, error) {
	if _, err := m.Get(br.Name); err != nil {
		return nil, errNoSuchBroker
	}
	m.brokers[br.Name] = br
	return br, nil
}

func (m *memStorageBroker) Delete(name string) error {
	if _, err := m.Get(name); err != nil {
		return errNoSuchBroker
	}
	delete(m.brokers, name)
	return nil
}
