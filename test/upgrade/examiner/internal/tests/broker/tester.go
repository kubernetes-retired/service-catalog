package broker

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scClientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

type tester struct {
	common
	c         scClientset.ServicecatalogV1beta1Interface
	namespace string
}

func newTester(cli ClientGetter, ns string) *tester {
	return &tester{
		c:         cli.ServiceCatalogClient().ServicecatalogV1beta1(),
		namespace: ns,
		common: common{
			sc:        cli.ServiceCatalogClient().ServicecatalogV1beta1(),
			namespace: ns,
		},
	}
}

func (t *tester) execute() error {
	klog.Info("Start test resources for ServiceBroker test")
	for _, fn := range []func() error{
		t.assertServiceBrokerIsReady,
		t.checkServiceClass,
		t.checkServicePlan,
		t.assertServiceInstanceIsReady,
		t.assertServiceBindingIsReady,
		t.removeServiceBinding,
		t.removeServiceInstance,
		t.unregisterServiceBroker,
	} {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *tester) assertServiceBrokerIsReady() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		broker, err := t.sc.ServiceBrokers(t.namespace).Get(serviceBrokerName, metav1.GetOptions{})
		if apiErr.IsNotFound(err) {
			klog.Infof("ServiceBroker %q not exist", serviceBrokerName)
			return false, nil
		}
		if err != nil {
			return false, err
		}

		condition := v1beta1.ServiceBrokerCondition{
			Type:    v1beta1.ServiceBrokerConditionReady,
			Status:  v1beta1.ConditionTrue,
			Message: successFetchedCatalogMessage,
		}
		for _, cond := range broker.Status.Conditions {
			if condition.Type == cond.Type && condition.Status == cond.Status && condition.Message == cond.Message {
				klog.Info("ServiceBroker is in ready state")
				return true, nil
			}
			klog.Infof("ServiceBroker is not ready, condition: Type: %q, Status: %q, Reason: %q", cond.Type, cond.Status, cond.Message)
		}

		return false, nil
	})
}

func (t *tester) removeServiceBinding() error {
	exist, err := t.serviceBindingExist()
	if err != nil {
		return errors.Wrap(err, "failed during fetching ServiceBinding")
	}
	if !exist {
		return nil
	}
	if err := t.deleteServiceBinding(); err != nil {
		return errors.Wrap(err, "failed during removing ServiceBinding")
	}
	if err := t.assertServiceBindingIsRemoved(); err != nil {
		return errors.Wrap(err, "failed during asserting ServiceBinding is removed")
	}
	return nil
}

func (t *tester) removeServiceInstance() error {
	exist, err := t.serviceInstanceExist()
	if err != nil {
		return errors.Wrap(err, "failed during fetching ServiceInstance")
	}
	if !exist {
		return nil
	}
	// remove `removeServiceInstanceFinalizer` method if TestBroker will be fixed and
	// will handle ServiceInstance delete operation
	// for now BrokerTest failed and ServiceInstance has deprovisioning false status
	// service patch is available on https://github.com/kubernetes-sigs/service-catalog/pull/2656
	// method can be removed when PR will be merged and TestBroker will be in version > 0.2.1
	if err := t.removeServiceInstanceFinalizer(); err != nil {
		return errors.Wrap(err, "failed during removing ServiceInstance finalizers")
	}
	if err := t.deleteServiceInstance(); err != nil {
		return errors.Wrap(err, "failed during removing ServiceInstance")
	}
	if err := t.assertServiceInstanceIsRemoved(); err != nil {
		return errors.Wrap(err, "failed during asserting ServiceInstance is removed")
	}
	return nil
}

func (t *tester) unregisterServiceBroker() error {
	if err := t.deleteServiceBroker(); err != nil {
		return errors.Wrap(err, "failed during removing ServiceBroker")
	}
	return nil
}

func (t *tester) serviceBindingExist() (bool, error) {
	_, err := t.sc.ServiceBindings(t.namespace).Get(serviceBindingName, metav1.GetOptions{})
	if apiErr.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (t *tester) deleteServiceBinding() error {
	err := t.sc.ServiceBindings(t.namespace).Delete(serviceBindingName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (t *tester) assertServiceBindingIsRemoved() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		_, err = t.sc.ServiceBindings(t.namespace).Get(serviceBindingName, metav1.GetOptions{})
		if apiErr.IsNotFound(err) {
			klog.Infof("ServiceBinding %q not exist", serviceBindingName)
			return true, nil
		}
		if err != nil {
			return false, err
		}

		return false, nil
	})
}

func (t *tester) serviceInstanceExist() (bool, error) {
	_, err := t.sc.ServiceInstances(t.namespace).Get(serviceInstanceName, metav1.GetOptions{})
	if apiErr.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (t *tester) removeServiceInstanceFinalizer() error {
	instance, err := t.sc.ServiceInstances(t.namespace).Get(serviceInstanceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	toUpdate := instance.DeepCopy()
	toUpdate.Finalizers = nil

	_, err = t.sc.ServiceInstances(toUpdate.Namespace).Update(toUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (t *tester) deleteServiceInstance() error {
	err := t.sc.ServiceInstances(t.namespace).Delete(serviceInstanceName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (t *tester) assertServiceInstanceIsRemoved() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		_, err = t.sc.ServiceInstances(t.namespace).Get(serviceInstanceName, metav1.GetOptions{})
		if apiErr.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		return false, nil
	})
}

func (t *tester) deleteServiceBroker() error {
	err := t.sc.ServiceBrokers(t.namespace).Delete(serviceBrokerName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
