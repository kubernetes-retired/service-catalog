package clusterbroker

import (
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"k8s.io/klog"
	"time"
)

const (
	clusterServiceBrokerName         = "test-broker"
	successFetchedCatalogMessage     = "Successfully fetched catalog entries from broker."
	amountOfClusterServiceClasses    = 24
	amountOfClusterServicePlans      = 25
	serviceInstanceName              = "test-instance"
	successProvisionMessage          = "The instance was provisioned successfully"
	serviceBindingName               = "test-binding"
	successInjectedBindResultMessage = "Injected bind result"

	waitInterval    = 1 * time.Second
	timeoutInterval = 20 * time.Second
)

// ClientGetter is an interface to represent structs return kubernetes clientset
type ClientGetter interface {
	ServiceCatalogClient() sc.Interface
}

// TestBroker represents upgrade test for ClusterServiceBroker
type TestBroker struct {
	client ClientGetter
}

// NewTestBroker is constructor for TestBroker
func NewTestBroker(cli ClientGetter) *TestBroker {
	return &TestBroker{cli}
}

// CreateResources prepares resources for upgrade test for ClusterServiceBroker
func (tb *TestBroker) CreateResources(stop <-chan struct{}, namespace string) error {
	c := newCreator(tb.client, namespace)

	klog.Info("Start creation process")
	return c.execute()
}

// TestResources executes test for ClusterServiceBroker and clean resource after finish
func (tb *TestBroker) TestResources(stop <-chan struct{}, namespace string) error {
	c := newTester(tb.client, namespace)

	klog.Info("Start test process")
	return c.execute()
}
