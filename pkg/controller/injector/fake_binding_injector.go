package injector

import (
	"errors"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi"
)

var (
	errNoSuchBinding = errors.New("no such binding")
)

// Fake is a fake implementation of a BindingInjector, intended for use in unit tests
type Fake struct {
	Injected map[*servicecatalog.Binding]*brokerapi.Credential
}

// NewFake creates a new Fake injector
func NewFake() *Fake {
	return &Fake{
		Injected: make(map[*servicecatalog.Binding]*brokerapi.Credential),
	}
}

// Inject records b and c in f.Injected and returns nil. This function is not concurrency-safe
func (f *Fake) Inject(b *servicecatalog.Binding, c *brokerapi.Credential) error {
	f.Injected[b] = c
	return nil
}

// Uninject returns an error if b doesn't exist in f.Injected. Otherwise, removes b
// from f.Injected and returns nil. This function is not concurrency-safe
func (f *Fake) Uninject(b *servicecatalog.Binding) error {
	_, ok := f.Injected[b]
	if !ok {
		return errNoSuchBinding
	}
	delete(f.Injected, b)
	return nil
}
