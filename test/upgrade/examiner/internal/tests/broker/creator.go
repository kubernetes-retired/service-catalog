package broker

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scClientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

type creator struct {
	common
	sc        scClientset.ServicecatalogV1beta1Interface
	namespace string
}

func newCreator(cli ClientGetter, ns string) *creator {
	return &creator{
		sc:        cli.ServiceCatalogClient().ServicecatalogV1beta1(),
		namespace: ns,
		common: common{
			sc:        cli.ServiceCatalogClient().ServicecatalogV1beta1(),
			namespace: ns,
		},
	}
}

func (c *creator) execute() error {
	klog.Info("Start prepare resources for ServiceBroker test")
	for _, fn := range []func() error{
		c.registerServiceBroker,
		c.checkServiceClass,
		c.checkServicePlan,
		c.createServiceInstance,
		c.createServiceBinding,
	} {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *creator) registerServiceBroker() error {
	klog.Infof("Create ServiceBroker %q", serviceBrokerName)
	if err := c.createServiceBroker(); err != nil {
		return errors.Wrap(err, "failed during creating ServiceBroker")
	}

	return nil
}

func (c *creator) createServiceBroker() error {
	_, err := c.sc.ServiceBrokers(c.namespace).Create(&v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceBrokerName,
			Namespace: c.namespace,
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: "http://test-broker-test-broker.test-broker.svc.cluster.local",
			},
		},
	})

	return err
}

func (c *creator) createServiceInstance() error {
	klog.Info("Create ServiceInstance")
	if err := c.createDefaultServiceInstance(); err != nil {
		return errors.Wrap(err, "failed during creating ServiceInstance")
	}

	klog.Info("Check ServiceInstance is ready")
	if err := c.assertServiceInstanceIsReady(); err != nil {
		return errors.Wrap(err, "failed during checking ServiceInstance conditions")
	}

	return nil
}

func (c *creator) createDefaultServiceInstance() error {
	_, err := c.sc.ServiceInstances(c.namespace).Create(&v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceInstanceName,
			Namespace: c.namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: "test-service-multiple-plans",
				ServicePlanExternalName:  "default",
			},
			Parameters: &runtime.RawExtension{
				Raw: []byte(`{ "param-1":"value-1", "param-2":"value-2" }`),
			},
		},
	})

	return err
}

func (c *creator) createServiceBinding() error {
	klog.Info("Create ServiceBinding")
	if err := c.createDefaultServiceBinding(); err != nil {
		return errors.Wrap(err, "failed during creating ServiceBinding")
	}

	klog.Info("Check ServiceBinding is ready")
	if err := c.assertServiceBindingIsReady(); err != nil {
		return errors.Wrap(err, "failed during checking ServiceBinding conditions")
	}

	return nil
}

func (c *creator) createDefaultServiceBinding() error {
	_, err := c.sc.ServiceBindings(c.namespace).Create(&v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceBindingName,
			Namespace: c.namespace,
		},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: v1beta1.LocalObjectReference{
				Name: serviceInstanceName,
			},
		},
	})

	return err
}
