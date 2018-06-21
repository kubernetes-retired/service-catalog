/*
Copyright 2017 The Kubernetes Authors.

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
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"

	corev1 "k8s.io/api/core/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	informers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1beta1"
	listers "github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1"
	scfeatures "github.com/kubernetes-incubator/service-catalog/pkg/features"
	"github.com/kubernetes-incubator/service-catalog/pkg/filter"
	"github.com/kubernetes-incubator/service-catalog/pkg/pretty"
)

const (
	// maxRetries is the number of times a resource add/update will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a resource is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
	// pollingStartInterval is the initial interval to use when polling async OSB operations.
	pollingStartInterval = 1 * time.Second

	// ContextProfilePlatformKubernetes is the platform name sent in the OSB
	// ContextProfile for requests coming from Kubernetes.
	ContextProfilePlatformKubernetes string = "kubernetes"
	// DefaultClusterIDConfigMapName is the k8s name that the clusterid configmap will have
	DefaultClusterIDConfigMapName string = "cluster-info"
	// DefaultClusterIDConfigMapNamespace is the k8s namespace that the clusterid configmap will be stored in.
	DefaultClusterIDConfigMapNamespace string = "default"
)

// NewController returns a new Open Service Broker catalog controller.
func NewController(
	kubeClient kubernetes.Interface,
	serviceCatalogClient servicecatalogclientset.ServicecatalogV1beta1Interface,
	clusterServiceBrokerInformer informers.ClusterServiceBrokerInformer,
	serviceBrokerInformer informers.ServiceBrokerInformer,
	clusterServiceClassInformer informers.ClusterServiceClassInformer,
	serviceClassInformer informers.ServiceClassInformer,
	instanceInformer informers.ServiceInstanceInformer,
	bindingInformer informers.ServiceBindingInformer,
	clusterServicePlanInformer informers.ClusterServicePlanInformer,
	servicePlanInformer informers.ServicePlanInformer,
	brokerClientCreateFunc osb.CreateFunc,
	brokerRelistInterval time.Duration,
	osbAPIPreferredVersion string,
	recorder record.EventRecorder,
	reconciliationRetryDuration time.Duration,
	operationPollingMaximumBackoffDuration time.Duration,
	clusterIDConfigMapName string,
	clusterIDConfigMapNamespace string,
) (Controller, error) {
	controller := &controller{
		kubeClient:                  kubeClient,
		serviceCatalogClient:        serviceCatalogClient,
		brokerClientCreateFunc:      brokerClientCreateFunc,
		brokerRelistInterval:        brokerRelistInterval,
		OSBAPIPreferredVersion:      osbAPIPreferredVersion,
		recorder:                    recorder,
		reconciliationRetryDuration: reconciliationRetryDuration,
		clusterServiceBrokerQueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster-service-broker"),
		serviceBrokerQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-broker"),
		clusterServiceClassQueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster-service-class"),
		serviceClassQueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-class"),
		clusterServicePlanQueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "cluster-service-plan"),
		servicePlanQueue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-plan"),
		instanceQueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-instance"),
		bindingQueue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "service-binding"),
		instancePollingQueue:        workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(pollingStartInterval, operationPollingMaximumBackoffDuration), "instance-poller"),
		bindingPollingQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(pollingStartInterval, operationPollingMaximumBackoffDuration), "binding-poller"),
		clusterIDConfigMapName:      clusterIDConfigMapName,
		clusterIDConfigMapNamespace: clusterIDConfigMapNamespace,
	}

	controller.clusterServiceBrokerLister = clusterServiceBrokerInformer.Lister()
	clusterServiceBrokerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.clusterServiceBrokerAdd,
		UpdateFunc: controller.clusterServiceBrokerUpdate,
		DeleteFunc: controller.clusterServiceBrokerDelete,
	})

	controller.clusterServiceClassLister = clusterServiceClassInformer.Lister()
	clusterServiceClassInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.clusterServiceClassAdd,
		UpdateFunc: controller.clusterServiceClassUpdate,
		DeleteFunc: controller.clusterServiceClassDelete,
	})

	controller.clusterServicePlanLister = clusterServicePlanInformer.Lister()
	clusterServicePlanInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.clusterServicePlanAdd,
		UpdateFunc: controller.clusterServicePlanUpdate,
		DeleteFunc: controller.clusterServicePlanDelete,
	})

	controller.instanceLister = instanceInformer.Lister()
	instanceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.instanceAdd,
		UpdateFunc: controller.instanceUpdate,
		DeleteFunc: controller.instanceDelete,
	})

	controller.bindingLister = bindingInformer.Lister()
	bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.bindingAdd,
		UpdateFunc: controller.bindingUpdate,
		DeleteFunc: controller.bindingDelete,
	})

	if utilfeature.DefaultFeatureGate.Enabled(scfeatures.NamespacedServiceBroker) {
		controller.serviceBrokerLister = serviceBrokerInformer.Lister()
		serviceBrokerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.serviceBrokerAdd,
			UpdateFunc: controller.serviceBrokerUpdate,
			DeleteFunc: controller.serviceBrokerDelete,
		})
		controller.serviceClassLister = serviceClassInformer.Lister()
		serviceClassInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.serviceClassAdd,
			UpdateFunc: controller.serviceClassUpdate,
			DeleteFunc: controller.serviceClassDelete,
		})
		controller.servicePlanLister = servicePlanInformer.Lister()
		//servicePlanInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		//AddFunc:    controller.servicePlanAdd,
		//UpdateFunc: controller.servicePlanUpdate,
		//DeleteFunc: controller.servicePlanDelete,
		//})
	}

	return controller, nil
}

// Controller describes a controller that backs the service catalog API for
// Open Service Broker compliant Brokers.
type Controller interface {
	// Run runs the controller until the given stop channel can be read from.
	// workers specifies the number of goroutines, per resource, processing work
	// from the resource workqueues
	Run(workers int, stopCh <-chan struct{})
}

// controller is a concrete Controller.
type controller struct {
	kubeClient                  kubernetes.Interface
	serviceCatalogClient        servicecatalogclientset.ServicecatalogV1beta1Interface
	brokerClientCreateFunc      osb.CreateFunc
	clusterServiceBrokerLister  listers.ClusterServiceBrokerLister
	serviceBrokerLister         listers.ServiceBrokerLister
	clusterServiceClassLister   listers.ClusterServiceClassLister
	serviceClassLister          listers.ServiceClassLister
	instanceLister              listers.ServiceInstanceLister
	bindingLister               listers.ServiceBindingLister
	clusterServicePlanLister    listers.ClusterServicePlanLister
	servicePlanLister           listers.ServicePlanLister
	brokerRelistInterval        time.Duration
	OSBAPIPreferredVersion      string
	recorder                    record.EventRecorder
	reconciliationRetryDuration time.Duration
	clusterServiceBrokerQueue   workqueue.RateLimitingInterface
	serviceBrokerQueue          workqueue.RateLimitingInterface
	clusterServiceClassQueue    workqueue.RateLimitingInterface
	serviceClassQueue           workqueue.RateLimitingInterface
	clusterServicePlanQueue     workqueue.RateLimitingInterface
	servicePlanQueue            workqueue.RateLimitingInterface
	instanceQueue               workqueue.RateLimitingInterface
	bindingQueue                workqueue.RateLimitingInterface
	instancePollingQueue        workqueue.RateLimitingInterface
	bindingPollingQueue         workqueue.RateLimitingInterface
	// clusterIDConfigMapName is the k8s name that the clusterid
	// configmap will have.
	clusterIDConfigMapName string
	// clusterIDConfigMapNamespace is the k8s namespace that the
	// clusterid configmap will be stored in.
	clusterIDConfigMapNamespace string
	// clusterID holds the current value. If a configmap to hold
	// this value does not exist, it will be created with this
	// value. If there is a configmap with a different value, it
	// will be reconciled to become the value in the configmap.
	clusterID string
	// clusterIDLock protects access to clusterID between the
	// monitor writing the value from the configmap, and any
	// readers passing the clusterID to a broker.
	clusterIDLock sync.RWMutex
}

// Run runs the controller until the given stop channel can be read from.
func (c *controller) Run(workers int, stopCh <-chan struct{}) {
	defer runtimeutil.HandleCrash()

	glog.Info("Starting service-catalog controller")

	var waitGroup sync.WaitGroup

	for i := 0; i < workers; i++ {
		createWorker(c.clusterServiceBrokerQueue, "ClusterServiceBroker", maxRetries, true, c.reconcileClusterServiceBrokerKey, stopCh, &waitGroup)
		createWorker(c.clusterServiceClassQueue, "ClusterServiceClass", maxRetries, true, c.reconcileClusterServiceClassKey, stopCh, &waitGroup)
		createWorker(c.clusterServicePlanQueue, "ClusterServicePlan", maxRetries, true, c.reconcileClusterServicePlanKey, stopCh, &waitGroup)
		createWorker(c.instanceQueue, "ServiceInstance", maxRetries, true, c.reconcileServiceInstanceKey, stopCh, &waitGroup)
		createWorker(c.bindingQueue, "ServiceBinding", maxRetries, true, c.reconcileServiceBindingKey, stopCh, &waitGroup)
		createWorker(c.instancePollingQueue, "InstancePoller", maxRetries, false, c.requeueServiceInstanceForPoll, stopCh, &waitGroup)

		if utilfeature.DefaultFeatureGate.Enabled(scfeatures.NamespacedServiceBroker) {
			createWorker(c.serviceBrokerQueue, "ServiceBroker", maxRetries, true, c.reconcileServiceBrokerKey, stopCh, &waitGroup)
			createWorker(c.serviceClassQueue, "ServiceClass", maxRetries, true, c.reconcileServiceClassKey, stopCh, &waitGroup)
		}

		if utilfeature.DefaultFeatureGate.Enabled(scfeatures.AsyncBindingOperations) {
			createWorker(c.bindingPollingQueue, "BindingPoller", maxRetries, false, c.requeueServiceBindingForPoll, stopCh, &waitGroup)
		}
	}

	// this creates a worker specifically for monitoring
	// configmaps, as we don't have the watching polling queue
	// infrastructure set up for one configmap. Instead this is a
	// simple polling based worker
	c.createConfigMapMonitorWorker(stopCh, &waitGroup)

	<-stopCh
	glog.Info("Shutting down service-catalog controller")

	c.clusterServiceBrokerQueue.ShutDown()
	c.clusterServiceClassQueue.ShutDown()
	c.clusterServicePlanQueue.ShutDown()
	c.instanceQueue.ShutDown()
	c.bindingQueue.ShutDown()
	c.instancePollingQueue.ShutDown()
	c.bindingPollingQueue.ShutDown()

	if utilfeature.DefaultFeatureGate.Enabled(scfeatures.NamespacedServiceBroker) {
		c.serviceBrokerQueue.ShutDown()
		c.serviceClassQueue.ShutDown()
	}

	waitGroup.Wait()
	glog.Info("Shutdown service-catalog controller")
}

// createWorker creates and runs a worker thread that just processes items in the
// specified queue. The worker will run until stopCh is closed. The worker will be
// added to the wait group when started and marked done when finished.
func createWorker(queue workqueue.RateLimitingInterface, resourceType string, maxRetries int, forgetAfterSuccess bool, reconciler func(key string) error, stopCh <-chan struct{}, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		wait.Until(worker(queue, resourceType, maxRetries, forgetAfterSuccess, reconciler), time.Second, stopCh)
		waitGroup.Done()
	}()
}

func (c *controller) createConfigMapMonitorWorker(stopCh <-chan struct{}, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		wait.Until(c.monitorConfigMap, 15*time.Second, stopCh)
		waitGroup.Done()
	}()
}

func (c *controller) monitorConfigMap() {
	// Cannot wait for the informer to push something into a queue.
	// What we're waiting on may never exist without us configuring
	// it, so we have to poll/ ask for it the first time to get it set.

	// Can we ask 'through' an informer? Is it a writeback cache? I
	// only ever want to monitor and be notified about one configmap
	// in a hardcoded place.
	glog.V(9).Info("cluster ID monitor loop enter")
	cm, err := c.kubeClient.CoreV1().ConfigMaps(c.clusterIDConfigMapNamespace).Get(c.clusterIDConfigMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		m := make(map[string]string)
		m["id"] = c.getClusterID()
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: c.clusterIDConfigMapName,
			},
			Data: m,
		}
		// if we fail to set the id,
		// it could be due to permissions
		// or due to being already set while we were trying
		if _, err := c.kubeClient.CoreV1().ConfigMaps(c.clusterIDConfigMapNamespace).Create(cm); err != nil {
			glog.Warningf("due to error %q, could not set clusterid configmap to %#v ", err, cm)
		}
	} else if err == nil {
		// cluster id exists and is set
		// get id out of cm
		if id := cm.Data["id"]; "" != id {
			c.setClusterID(id)
		} else {
			m := cm.Data
			if m == nil {
				m = make(map[string]string)
				cm.Data = m
			}
			m["id"] = c.getClusterID()
			c.kubeClient.CoreV1().ConfigMaps(c.clusterIDConfigMapNamespace).Update(cm)
		}
	} else { // some err we can't handle
		glog.V(4).Infof("error getting the cluster info configmap: %q", err)
	}
	glog.V(9).Info("cluster ID monitor loop exit")
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// If reconciler returns an error, requeue the item up to maxRetries before giving up.
// It enforces that the reconciler is never invoked concurrently with the same key.
// If forgetAfterSuccess is true, it will cause the queue to forget the item should reconciliation
// have no error.
func worker(queue workqueue.RateLimitingInterface, resourceType string, maxRetries int, forgetAfterSuccess bool, reconciler func(key string) error) func() {
	return func() {
		exit := false
		for !exit {
			exit = func() bool {
				key, quit := queue.Get()
				if quit {
					return true
				}
				defer queue.Done(key)

				err := reconciler(key.(string))
				if err == nil {
					if forgetAfterSuccess {
						queue.Forget(key)
					}
					return false
				}

				numRequeues := queue.NumRequeues(key)
				if numRequeues < maxRetries {
					glog.V(4).Infof("Error syncing %s %v (retry: %d/%d): %v", resourceType, key, numRequeues, maxRetries, err)
					queue.AddRateLimited(key)
					return false
				}

				glog.V(4).Infof("Dropping %s %q out of the queue: %v", resourceType, key, err)
				queue.Forget(key)
				return false
			}()
		}
	}
}

// operationError is a user-facing error that can be easily embedded in a
// resource's Condition.
type operationError struct {
	reason  string
	message string
}

func (e *operationError) Error() string { return e.message }

// getClusterServiceClassPlanAndClusterServiceBroker is a sequence of operations that's done in couple of
// places so this method fetches the Service Class, Service Plan and creates
// a brokerClient to use for that method given an ServiceInstance.
// The ClusterServicePlan returned will be nil if the ClusterServicePlanRef
// is nil. This will happen when deleting a ServiceInstance that previously
// had an update to a non-existent plan.
func (c *controller) getClusterServiceClassPlanAndClusterServiceBroker(instance *v1beta1.ServiceInstance) (*v1beta1.ClusterServiceClass, *v1beta1.ClusterServicePlan, string, osb.Client, error) {
	serviceClass, brokerName, brokerClient, err := c.getClusterServiceClassAndClusterServiceBroker(instance)
	if err != nil {
		return nil, nil, "", nil, err
	}

	var servicePlan *v1beta1.ClusterServicePlan
	if instance.Spec.ClusterServicePlanRef != nil {
		var err error
		servicePlan, err = c.clusterServicePlanLister.Get(instance.Spec.ClusterServicePlanRef.Name)
		if nil != err {
			return nil, nil, "", nil, &operationError{
				reason: errorNonexistentClusterServicePlanReason,
				message: fmt.Sprintf(
					"The instance references a non-existent ClusterServicePlan %q - %v",
					instance.Spec.ClusterServicePlanRef.Name, instance.Spec.PlanReference,
				),
			}
		}
	}
	return serviceClass, servicePlan, brokerName, brokerClient, nil
}

func (c *controller) getServiceClassPlanAndServiceBroker(instance *v1beta1.ServiceInstance) (*v1beta1.ServiceClass, *v1beta1.ServicePlan, string, osb.Client, error) {
	serviceClass, brokerName, brokerClient, err := c.getServiceClassAndServiceBroker(instance)
	if err != nil {
		return nil, nil, "", nil, err
	}

	var servicePlan *v1beta1.ServicePlan
	if instance.Spec.ServicePlanRef != nil {
		var err error
		servicePlan, err = c.servicePlanLister.ServicePlans(instance.Namespace).Get(instance.Spec.ServicePlanRef.Name)
		if nil != err {
			return nil, nil, "", nil, &operationError{
				reason: errorNonexistentServicePlanReason,
				message: fmt.Sprintf(
					"The instance references a non-existent ServicePlan %q - %v",
					instance.Spec.ServicePlanRef.Name, instance.Spec.PlanReference,
				),
			}
		}
	}
	return serviceClass, servicePlan, brokerName, brokerClient, nil
}

// getClusterServiceClassAndClusterServiceBroker is a sequence of operations that's done in couple of
// places so this method fetches the Service Class and creates
// a brokerClient to use for that method given an ServiceInstance.
func (c *controller) getClusterServiceClassAndClusterServiceBroker(instance *v1beta1.ServiceInstance) (*v1beta1.ClusterServiceClass, string, osb.Client, error) {
	pcb := pretty.NewInstanceContextBuilder(instance)
	serviceClass, err := c.clusterServiceClassLister.Get(instance.Spec.ClusterServiceClassRef.Name)
	if err != nil {
		return nil, "", nil, &operationError{
			reason: errorNonexistentClusterServiceClassReason,
			message: fmt.Sprintf(
				"The instance references a non-existent ClusterServiceClass (K8S: %q ExternalName: %q)",
				instance.Spec.ClusterServiceClassRef.Name, instance.Spec.ClusterServiceClassExternalName,
			),
		}
	}

	broker, err := c.clusterServiceBrokerLister.Get(serviceClass.Spec.ClusterServiceBrokerName)
	if err != nil {
		return nil, "", nil, &operationError{
			reason: errorNonexistentClusterServiceBrokerReason,
			message: fmt.Sprintf(
				"The instance references a non-existent broker %q",
				serviceClass.Spec.ClusterServiceBrokerName,
			),
		}

	}

	authConfig, err := getAuthCredentialsFromClusterServiceBroker(c.kubeClient, broker)
	if err != nil {
		return nil, "", nil, &operationError{
			reason: errorAuthCredentialsReason,
			message: fmt.Sprintf(
				"Error getting broker auth credentials for broker %q: %s",
				broker.Name, err,
			),
		}
	}

	clientConfig := NewClientConfigurationForBroker(broker.ObjectMeta, &broker.Spec.CommonServiceBrokerSpec, authConfig)
	glog.V(4).Info(pcb.Messagef("Creating client for ClusterServiceBroker %v, URL: %v", broker.Name, broker.Spec.URL))
	brokerClient, err := c.brokerClientCreateFunc(clientConfig)
	if err != nil {
		return nil, "", nil, err
	}

	return serviceClass, broker.Name, brokerClient, nil
}

// getServiceClassAndServiceBroker is a sequence of operations that's done in couple of
// places so this method fetches the Service Class and creates
// a brokerClient to use for that method given a ServiceInstance.
func (c *controller) getServiceClassAndServiceBroker(instance *v1beta1.ServiceInstance) (*v1beta1.ServiceClass, string, osb.Client, error) {
	pcb := pretty.NewContextBuilder(pretty.ServiceInstance, instance.Namespace, instance.Name, "")
	serviceClass, err := c.serviceClassLister.ServiceClasses(instance.Namespace).Get(instance.Spec.ServiceClassRef.Name)
	if err != nil {
		return nil, "", nil, &operationError{
			reason: errorNonexistentServiceClassReason,
			message: fmt.Sprintf(
				"The instance references a non-existent ServiceClass (K8S: %q ExternalName: %q)",
				instance.Spec.ServiceClassRef.Name, instance.Spec.ServiceClassExternalName,
			),
		}
	}

	broker, err := c.serviceBrokerLister.ServiceBrokers(instance.Namespace).Get(serviceClass.Spec.ServiceBrokerName)
	if err != nil {
		return nil, "", nil, &operationError{
			reason: errorNonexistentServiceBrokerReason,
			message: fmt.Sprintf(
				"The instance references a non-existent broker %q",
				serviceClass.Spec.ServiceBrokerName,
			),
		}

	}

	authConfig, err := getAuthCredentialsFromServiceBroker(c.kubeClient, broker)
	if err != nil {
		return nil, "", nil, &operationError{
			reason: errorAuthCredentialsReason,
			message: fmt.Sprintf(
				"Error getting broker auth credentials for broker %q: %s",
				broker.Name, err,
			),
		}
	}

	clientConfig := NewClientConfigurationForBroker(broker.ObjectMeta, &broker.Spec.CommonServiceBrokerSpec, authConfig)
	glog.V(4).Info(pcb.Messagef("Creating client for ServiceBroker %v, URL: %v", broker.Name, broker.Spec.URL))
	brokerClient, err := c.brokerClientCreateFunc(clientConfig)
	if err != nil {
		return nil, "", nil, err
	}

	return serviceClass, broker.Name, brokerClient, nil
}

// getClusterServiceClassPlanAndClusterServiceBrokerForServiceBinding is a sequence of operations that's
// done to validate service plan, service class exist, and handles creating
// a brokerclient to use for a given ServiceInstance.
// Sets ClusterServiceClassRef and/or ClusterServicePlanRef if they haven't been already set.
func (c *controller) getClusterServiceClassPlanAndClusterServiceBrokerForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding) (*v1beta1.ClusterServiceClass, *v1beta1.ClusterServicePlan, string, osb.Client, error) {
	serviceClass, serviceBrokerName, osbClient, err := c.getClusterServiceClassAndClusterServiceBrokerForServiceBinding(instance, binding)
	if err != nil {
		return nil, nil, "", nil, err
	}
	servicePlan, err := c.getClusterServicePlanForServiceBinding(instance, binding, serviceClass)
	if err != nil {
		return nil, nil, "", nil, err
	}

	return serviceClass, servicePlan, serviceBrokerName, osbClient, nil
}

func (c *controller) getClusterServiceClassAndClusterServiceBrokerForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding) (*v1beta1.ClusterServiceClass, string, osb.Client, error) {
	serviceClass, err := c.getClusterServiceClassForServiceBinding(instance, binding)
	if err != nil {
		return nil, "", nil, err
	}

	serviceBroker, err := c.getClusterServiceBrokerForServiceBinding(instance, binding, serviceClass)
	if err != nil {
		return nil, "", nil, err
	}

	//osbClient, err := c.getBrokerClientForServiceBinding(instance, binding, serviceBroker)
	osbClient, err := c.getBrokerClientForServiceBinding(instance, binding)
	if err != nil {
		return nil, "", nil, err
	}

	return serviceClass, serviceBroker.Name, osbClient, nil
}

func (c *controller) getClusterServiceClassForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding) (*v1beta1.ClusterServiceClass, error) {
	pcb := pretty.NewInstanceContextBuilder(instance)
	serviceClass, err := c.clusterServiceClassLister.Get(instance.Spec.ClusterServiceClassRef.Name)
	if err != nil {
		s := fmt.Sprintf(
			"References a non-existent ClusterServiceClass %q - %c",
			instance.Spec.ClusterServiceClassRef.Name, instance.Spec.PlanReference,
		)
		glog.Warning(pcb.Message(s))
		c.updateServiceBindingCondition(
			binding,
			v1beta1.ServiceBindingConditionReady,
			v1beta1.ConditionFalse,
			errorNonexistentClusterServiceClassReason,
			"The binding references a ClusterServiceClass that does not exist. "+s,
		)
		c.recorder.Event(binding, corev1.EventTypeWarning, errorNonexistentClusterServiceClassMessage, s)
		return nil, err
	}
	return serviceClass, nil
}

func (c *controller) getClusterServicePlanForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding, serviceClass *v1beta1.ClusterServiceClass) (*v1beta1.ClusterServicePlan, error) {
	pcb := pretty.NewInstanceContextBuilder(instance)
	servicePlan, err := c.clusterServicePlanLister.Get(instance.Spec.ClusterServicePlanRef.Name)
	if nil != err {
		s := fmt.Sprintf(
			"References a non-existent ClusterServicePlan %q - %v",
			instance.Spec.ClusterServicePlanRef.Name, instance.Spec.PlanReference,
		)
		glog.Warning(pcb.Message(s))
		c.updateServiceBindingCondition(
			binding,
			v1beta1.ServiceBindingConditionReady,
			v1beta1.ConditionFalse,
			errorNonexistentClusterServicePlanReason,
			"The ServiceBinding references an ServiceInstance which references ClusterServicePlan that does not exist. "+s,
		)
		c.recorder.Event(binding, corev1.EventTypeWarning, errorNonexistentClusterServicePlanReason, s)
		return nil, fmt.Errorf(s)
	}
	return servicePlan, nil
}

func (c *controller) getClusterServiceBrokerForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding, serviceClass *v1beta1.ClusterServiceClass) (*v1beta1.ClusterServiceBroker, error) {
	pcb := pretty.NewInstanceContextBuilder(instance)

	broker, err := c.clusterServiceBrokerLister.Get(serviceClass.Spec.ClusterServiceBrokerName)
	if err != nil {
		s := fmt.Sprintf("References a non-existent ClusterServiceBroker %q", serviceClass.Spec.ClusterServiceBrokerName)
		glog.Warning(pcb.Message(s))
		c.updateServiceBindingCondition(
			binding,
			v1beta1.ServiceBindingConditionReady,
			v1beta1.ConditionFalse,
			errorNonexistentClusterServiceBrokerReason,
			"The binding references a ClusterServiceBroker that does not exist. "+s,
		)
		c.recorder.Event(binding, corev1.EventTypeWarning, errorNonexistentClusterServiceBrokerReason, s)
		return nil, err
	}
	return broker, nil
}

/*
func (c *controller) getBrokerClientForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding, broker *v1beta1.ClusterServiceBroker) (osb.Client, error) {
	pcb := pretty.NewInstanceContextBuilder(instance)
	authConfig, err := getAuthCredentialsFromClusterServiceBroker(c.kubeClient, broker)
	if err != nil {
		s := fmt.Sprintf("Error getting broker auth credentials for broker %q: %s", broker.Name, err)
		glog.Warning(pcb.Message(s))
		c.updateServiceBindingCondition(
			binding,
			v1beta1.ServiceBindingConditionReady,
			v1beta1.ConditionFalse,
			errorAuthCredentialsReason,
			"Error getting auth credentials. "+s,
		)
		c.recorder.Event(binding, corev1.EventTypeWarning, errorAuthCredentialsReason, s)
		return nil, err
	}

	clientConfig := NewClientConfigurationForBroker(broker.ObjectMeta, &broker.Spec.CommonServiceBrokerSpec, authConfig)

	glog.V(4).Infof("Creating client for ClusterServiceBroker %v, URL: %v", broker.Name, broker.Spec.URL)
	brokerClient, err := c.brokerClientCreateFunc(clientConfig)
	if err != nil {
		return nil, err
	}

	return brokerClient, nil
}

/*/
func (c *controller) getBrokerClientForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding) (osb.Client, error) {

	var brokerClient osb.Client

	if instance.Spec.ClusterServiceClassSpecified() {

		serviceClass, err := c.getClusterServiceClassForServiceBinding(instance, binding)
		if err != nil {
			return nil, err
		}

		broker, err := c.getClusterServiceBrokerForServiceBinding(instance, binding, serviceClass)
		if err != nil {
			return nil, err
		}

		pcb := pretty.NewInstanceContextBuilder(instance)
		authConfig, err := getAuthCredentialsFromClusterServiceBroker(c.kubeClient, broker)
		if err != nil {
			s := fmt.Sprintf("Error getting broker auth credentials for broker %q: %s", broker.Name, err)
			glog.Warning(pcb.Message(s))
			c.updateServiceBindingCondition(
				binding,
				v1beta1.ServiceBindingConditionReady,
				v1beta1.ConditionFalse,
				errorAuthCredentialsReason,
				"Error getting auth credentials. "+s,
			)
			c.recorder.Event(binding, corev1.EventTypeWarning, errorAuthCredentialsReason, s)
			return nil, err
		}

		clientConfig := NewClientConfigurationForBroker(broker.ObjectMeta, &broker.Spec.CommonServiceBrokerSpec, authConfig)

		glog.V(4).Infof("Creating client for ClusterServiceBroker %v, URL: %v", broker.Name, broker.Spec.URL)
		brokerClient, err = c.brokerClientCreateFunc(clientConfig)
		if err != nil {
			return nil, err
		}

	} else if instance.Spec.ServiceClassSpecified() {

		serviceClass, err := c.getServiceClassForServiceBinding(instance, binding)
		if err != nil {
			return nil, err
		}

		broker, err := c.getServiceBrokerForServiceBinding(instance, binding, serviceClass)
		if err != nil {
			return nil, err
		}

		pcb := pretty.NewInstanceContextBuilder(instance)
		authConfig, err := getAuthCredentialsFromServiceBroker(c.kubeClient, broker)
		if err != nil {
			s := fmt.Sprintf("Error getting broker auth credentials for broker %q: %s", broker.Name, err)
			glog.Warning(pcb.Message(s))
			c.updateServiceBindingCondition(
				binding,
				v1beta1.ServiceBindingConditionReady,
				v1beta1.ConditionFalse,
				errorAuthCredentialsReason,
				"Error getting auth credentials. "+s,
			)
			c.recorder.Event(binding, corev1.EventTypeWarning, errorAuthCredentialsReason, s)
			return nil, err
		}

		clientConfig := NewClientConfigurationForBroker(broker.ObjectMeta, &broker.Spec.CommonServiceBrokerSpec, authConfig)

		glog.V(4).Infof("Creating client for ClusterServiceBroker %v, URL: %v", broker.Name, broker.Spec.URL)
		brokerClient, err = c.brokerClientCreateFunc(clientConfig)
		if err != nil {
			return nil, err
		}
	}

	return brokerClient, nil
}

//*/

// Broker utility methods - move?
// getAuthCredentialsFromClusterServiceBroker returns the auth credentials, if any, or
// returns an error. If the AuthInfo field is nil, empty values are
// returned.
func getAuthCredentialsFromClusterServiceBroker(client kubernetes.Interface, broker *v1beta1.ClusterServiceBroker) (*osb.AuthConfig, error) {
	if broker.Spec.AuthInfo == nil {
		return nil, nil
	}

	authInfo := broker.Spec.AuthInfo
	if authInfo.Basic != nil {
		secretRef := authInfo.Basic.SecretRef
		secret, err := client.CoreV1().Secrets(secretRef.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		basicAuthConfig, err := getBasicAuthConfig(secret)
		if err != nil {
			return nil, err
		}
		return &osb.AuthConfig{
			BasicAuthConfig: basicAuthConfig,
		}, nil
	} else if authInfo.Bearer != nil {
		secretRef := authInfo.Bearer.SecretRef
		secret, err := client.CoreV1().Secrets(secretRef.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		bearerConfig, err := getBearerConfig(secret)
		if err != nil {
			return nil, err
		}
		return &osb.AuthConfig{
			BearerConfig: bearerConfig,
		}, nil
	}
	return nil, fmt.Errorf("empty auth info or unsupported auth mode: %s", authInfo)
}

// getAuthCredentialsFromServiceBroker returns the auth credentials, if any, or
// returns an error. If the AuthInfo field is nil, empty values are returned.
func getAuthCredentialsFromServiceBroker(client kubernetes.Interface, broker *v1beta1.ServiceBroker) (*osb.AuthConfig, error) {
	if broker.Spec.AuthInfo == nil {
		return nil, nil
	}

	authInfo := broker.Spec.AuthInfo
	if authInfo.Basic != nil {
		secretRef := authInfo.Basic.SecretRef
		secret, err := client.CoreV1().Secrets(broker.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		basicAuthConfig, err := getBasicAuthConfig(secret)
		if err != nil {
			return nil, err
		}
		return &osb.AuthConfig{
			BasicAuthConfig: basicAuthConfig,
		}, nil
	} else if authInfo.Bearer != nil {
		secretRef := authInfo.Bearer.SecretRef
		secret, err := client.CoreV1().Secrets(broker.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		bearerConfig, err := getBearerConfig(secret)
		if err != nil {
			return nil, err
		}
		return &osb.AuthConfig{
			BearerConfig: bearerConfig,
		}, nil
	}
	return nil, fmt.Errorf("empty auth info or unsupported auth mode: %s", authInfo)
}

func getBasicAuthConfig(secret *corev1.Secret) (*osb.BasicAuthConfig, error) {
	usernameBytes, ok := secret.Data["username"]
	if !ok {
		return nil, fmt.Errorf("auth secret didn't contain username")
	}

	passwordBytes, ok := secret.Data["password"]
	if !ok {
		return nil, fmt.Errorf("auth secret didn't contain password")
	}

	return &osb.BasicAuthConfig{
		Username: string(usernameBytes),
		Password: string(passwordBytes),
	}, nil
}

func getBearerConfig(secret *corev1.Secret) (*osb.BearerConfig, error) {
	tokenBytes, ok := secret.Data["token"]
	if !ok {
		return nil, fmt.Errorf("auth secret didn't contain token")
	}

	return &osb.BearerConfig{
		Token: string(tokenBytes),
	}, nil
}

// convertAndFilterCatalogToNamespacedTypes converts a service broker catalog
// into an array of ServiceClasses and an array of ServicePlans and filters
// these through the restrictions provided. The ServiceClasses and
// ServicePlans returned by this method are named in K8S with the OSB ID.
func convertAndFilterCatalogToNamespacedTypes(namespace string, in *osb.CatalogResponse, restrictions *v1beta1.CatalogRestrictions) ([]*v1beta1.ServiceClass, []*v1beta1.ServicePlan, error) {
	var predicate filter.Predicate
	var err error
	if restrictions != nil && len(restrictions.ServiceClass) > 0 {
		predicate, err = filter.CreatePredicate(restrictions.ServiceClass)
		if err != nil {
			return nil, nil, err
		}
	} else {
		predicate = filter.NewPredicate()
	}

	serviceClasses := []*v1beta1.ServiceClass(nil)
	servicePlans := []*v1beta1.ServicePlan(nil)
	for _, svc := range in.Services {
		serviceClass := &v1beta1.ServiceClass{
			Spec: v1beta1.ServiceClassSpec{
				CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
					Bindable:      svc.Bindable,
					PlanUpdatable: svc.PlanUpdatable != nil && *svc.PlanUpdatable,
					ExternalID:    svc.ID,
					ExternalName:  svc.Name,
					Tags:          svc.Tags,
					Description:   svc.Description,
					Requires:      svc.Requires,
				},
			},
		}

		if utilfeature.DefaultFeatureGate.Enabled(scfeatures.AsyncBindingOperations) {
			serviceClass.Spec.BindingRetrievable = svc.BindingsRetrievable
		}

		if svc.Metadata != nil {
			metadata, err := json.Marshal(svc.Metadata)
			if err != nil {
				err = fmt.Errorf("Failed to marshal metadata\n%+v\n %v", svc.Metadata, err)
				glog.Error(err)
				return nil, nil, err
			}
			serviceClass.Spec.ExternalMetadata = &runtime.RawExtension{Raw: metadata}
		}
		serviceClass.SetName(svc.ID)
		serviceClass.SetNamespace(namespace)

		// If this service class passes the predicate, process the plans for the class.
		if fields := v1beta1.ConvertServiceClassToProperties(serviceClass); predicate.Accepts(fields) {
			// set up the plans using the ServiceClass Name
			plans, err := convertServicePlans(namespace, svc.Plans, serviceClass.Name)
			if err != nil {
				return nil, nil, err
			}

			acceptedPlans, _, err := filterNamespacedServicePlans(restrictions, plans)
			if err != nil {
				return nil, nil, err
			}

			// If there are accepted plans, then append the class and all of the accepted plans to the master list.
			if len(acceptedPlans) > 0 {
				serviceClasses = append(serviceClasses, serviceClass)
				servicePlans = append(servicePlans, acceptedPlans...)
			}
		}
	}
	return serviceClasses, servicePlans, nil
}

// convertAndFilterCatalog converts a service broker catalog into an array of
// ClusterServiceClasses and an array of ClusterServicePlans and filters these
// through the restrictions provided. The ClusterServiceClasses and
// ClusterServicePlans returned by this method are named in K8S with the OSB ID.
func convertAndFilterCatalog(in *osb.CatalogResponse, restrictions *v1beta1.CatalogRestrictions) ([]*v1beta1.ClusterServiceClass, []*v1beta1.ClusterServicePlan, error) {
	var predicate filter.Predicate
	var err error
	if restrictions != nil && len(restrictions.ServiceClass) > 0 {
		predicate, err = filter.CreatePredicate(restrictions.ServiceClass)
		if err != nil {
			return nil, nil, err
		}
	} else {
		predicate = filter.NewPredicate()
	}

	serviceClasses := []*v1beta1.ClusterServiceClass(nil)
	servicePlans := []*v1beta1.ClusterServicePlan(nil)
	for _, svc := range in.Services {
		serviceClass := &v1beta1.ClusterServiceClass{
			Spec: v1beta1.ClusterServiceClassSpec{
				CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
					Bindable:      svc.Bindable,
					PlanUpdatable: svc.PlanUpdatable != nil && *svc.PlanUpdatable,
					ExternalID:    svc.ID,
					ExternalName:  svc.Name,
					Tags:          svc.Tags,
					Description:   svc.Description,
					Requires:      svc.Requires,
				},
			},
		}

		if utilfeature.DefaultFeatureGate.Enabled(scfeatures.AsyncBindingOperations) {
			serviceClass.Spec.BindingRetrievable = svc.BindingsRetrievable
		}

		if svc.Metadata != nil {
			metadata, err := json.Marshal(svc.Metadata)
			if err != nil {
				err = fmt.Errorf("Failed to marshal metadata\n%+v\n %v", svc.Metadata, err)
				glog.Error(err)
				return nil, nil, err
			}
			serviceClass.Spec.ExternalMetadata = &runtime.RawExtension{Raw: metadata}
		}
		serviceClass.SetName(svc.ID)

		// If this service class passes the predicate, process the plans for the class.
		if fields := v1beta1.ConvertClusterServiceClassToProperties(serviceClass); predicate.Accepts(fields) {
			// set up the plans using the ClusterServiceClass Name
			plans, err := convertClusterServicePlans(svc.Plans, serviceClass.Name)
			if err != nil {
				return nil, nil, err
			}

			acceptedPlans, _, err := filterServicePlans(restrictions, plans)
			if err != nil {
				return nil, nil, err
			}

			// If there are accepted plans, then append the class and all of the accepted plans to the master list.
			if len(acceptedPlans) > 0 {
				serviceClasses = append(serviceClasses, serviceClass)
				servicePlans = append(servicePlans, acceptedPlans...)
			}
		}
	}
	return serviceClasses, servicePlans, nil
}

func filterNamespacedServicePlans(restrictions *v1beta1.CatalogRestrictions, servicePlans []*v1beta1.ServicePlan) ([]*v1beta1.ServicePlan, []*v1beta1.ServicePlan, error) {
	var predicate filter.Predicate
	var err error
	if restrictions != nil && len(restrictions.ServicePlan) > 0 {
		predicate, err = filter.CreatePredicate(restrictions.ServicePlan)
		if err != nil {
			return nil, nil, err
		}
	} else {
		predicate = filter.NewPredicate()
	}

	// If the predicate is empty, all plans will pass. No need to run through the list.
	if predicate.Empty() {
		return servicePlans, []*v1beta1.ServicePlan(nil), nil
	}

	accepted := []*v1beta1.ServicePlan(nil)
	rejected := []*v1beta1.ServicePlan(nil)
	for _, sp := range servicePlans {
		fields := v1beta1.ConvertServicePlanToProperties(sp)
		if predicate.Accepts(fields) {
			accepted = append(accepted, sp)
		} else {
			rejected = append(rejected, sp)
		}
	}

	return accepted, rejected, nil
}

func filterServicePlans(restrictions *v1beta1.CatalogRestrictions, servicePlans []*v1beta1.ClusterServicePlan) ([]*v1beta1.ClusterServicePlan, []*v1beta1.ClusterServicePlan, error) {
	var predicate filter.Predicate
	var err error
	if restrictions != nil && len(restrictions.ServicePlan) > 0 {
		predicate, err = filter.CreatePredicate(restrictions.ServicePlan)
		if err != nil {
			return nil, nil, err
		}
	} else {
		predicate = filter.NewPredicate()
	}

	// If the predicate is empty, all plans will pass. No need to run through the list.
	if predicate.Empty() {
		return servicePlans, []*v1beta1.ClusterServicePlan(nil), nil
	}

	accepted := []*v1beta1.ClusterServicePlan(nil)
	rejected := []*v1beta1.ClusterServicePlan(nil)
	for _, sp := range servicePlans {
		fields := v1beta1.ConvertClusterServicePlanToProperties(sp)
		if predicate.Accepts(fields) {
			accepted = append(accepted, sp)
		} else {
			rejected = append(rejected, sp)
		}
	}

	return accepted, rejected, nil
}

func convertServicePlans(namespace string, plans []osb.Plan, serviceClassID string) ([]*v1beta1.ServicePlan, error) {
	if 0 == len(plans) {
		return nil, fmt.Errorf("ServiceClass (K8S: %q) must have at least one plan", serviceClassID)
	}
	servicePlans := make([]*v1beta1.ServicePlan, len(plans))
	for i, plan := range plans {
		servicePlan := &v1beta1.ServicePlan{
			Spec: v1beta1.ServicePlanSpec{
				CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
					ExternalName: plan.Name,
					ExternalID:   plan.ID,
					Free:         plan.Free != nil && *plan.Free,
					Description:  plan.Description,
				},
				ServiceClassRef: v1beta1.LocalObjectReference{Name: serviceClassID},
			},
		}
		servicePlans[i] = servicePlan
		servicePlan.SetName(plan.ID)
		servicePlan.SetNamespace(namespace)

		err := convertCommonServicePlan(plan, &servicePlan.Spec.CommonServicePlanSpec)
		if err != nil {
			return nil, err
		}
	}
	return servicePlans, nil
}

func convertCommonServicePlan(plan osb.Plan, commonServicePlanSpec *v1beta1.CommonServicePlanSpec) error {
	if plan.Bindable != nil {
		b := plan.Bindable
		commonServicePlanSpec.Bindable = b
	}

	if plan.Metadata != nil {
		metadata, err := json.Marshal(plan.Metadata)
		if err != nil {
			err = fmt.Errorf("Failed to marshal metadata\n%+v\n %v", plan.Metadata, err)
			glog.Error(err)
			return err
		}
		commonServicePlanSpec.ExternalMetadata = &runtime.RawExtension{Raw: metadata}
	}

	if schemas := plan.Schemas; schemas != nil {
		if instanceSchemas := schemas.ServiceInstance; instanceSchemas != nil {
			if instanceCreateSchema := instanceSchemas.Create; instanceCreateSchema != nil && instanceCreateSchema.Parameters != nil {
				schema, err := json.Marshal(instanceCreateSchema.Parameters)
				if err != nil {
					err = fmt.Errorf("Failed to marshal instance create schema \n%+v\n %v", instanceCreateSchema.Parameters, err)
					glog.Error(err)
					return err
				}
				commonServicePlanSpec.ServiceInstanceCreateParameterSchema = &runtime.RawExtension{Raw: schema}
			}
			if instanceUpdateSchema := instanceSchemas.Update; instanceUpdateSchema != nil && instanceUpdateSchema.Parameters != nil {
				schema, err := json.Marshal(instanceUpdateSchema.Parameters)
				if err != nil {
					err = fmt.Errorf("Failed to marshal instance update schema \n%+v\n %v", instanceUpdateSchema.Parameters, err)
					glog.Error(err)
					return err
				}
				commonServicePlanSpec.ServiceInstanceUpdateParameterSchema = &runtime.RawExtension{Raw: schema}
			}
		}
		if bindingSchemas := schemas.ServiceBinding; bindingSchemas != nil {
			if bindingCreateSchema := bindingSchemas.Create; bindingCreateSchema != nil {
				if bindingCreateSchema.Parameters != nil {
					schema, err := json.Marshal(bindingCreateSchema.Parameters)
					if err != nil {
						err = fmt.Errorf("Failed to marshal binding create schema \n%+v\n %v", bindingCreateSchema.Parameters, err)
						glog.Error(err)
						return err
					}
					commonServicePlanSpec.ServiceBindingCreateParameterSchema = &runtime.RawExtension{Raw: schema}
				}
				if utilfeature.DefaultFeatureGate.Enabled(scfeatures.ResponseSchema) && bindingCreateSchema.Response != nil {
					schema, err := json.Marshal(bindingCreateSchema.Response)
					if err != nil {
						err = fmt.Errorf("Failed to marshal binding create response schema \n%+v\n %v", bindingCreateSchema.Response, err)
						glog.Error(err)
						return err
					}
					commonServicePlanSpec.ServiceBindingCreateResponseSchema = &runtime.RawExtension{Raw: schema}
				}
			}
		}
	}
	return nil
}

func convertClusterServicePlans(plans []osb.Plan, serviceClassID string) ([]*v1beta1.ClusterServicePlan, error) {
	if 0 == len(plans) {
		return nil, fmt.Errorf("ClusterServiceClass (K8S: %q) must have at least one plan", serviceClassID)
	}
	servicePlans := make([]*v1beta1.ClusterServicePlan, len(plans))
	for i, plan := range plans {
		servicePlans[i] = &v1beta1.ClusterServicePlan{
			Spec: v1beta1.ClusterServicePlanSpec{
				CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
					ExternalName: plan.Name,
					ExternalID:   plan.ID,
					Free:         plan.Free != nil && *plan.Free,
					Description:  plan.Description,
				},
				ClusterServiceClassRef: v1beta1.ClusterObjectReference{Name: serviceClassID},
			},
		}
		servicePlans[i].SetName(plan.ID)

		if plan.Bindable != nil {
			b := *plan.Bindable
			servicePlans[i].Spec.Bindable = &b
		}

		if plan.Metadata != nil {
			metadata, err := json.Marshal(plan.Metadata)
			if err != nil {
				err = fmt.Errorf("Failed to marshal metadata\n%+v\n %v", plan.Metadata, err)
				glog.Error(err)
				return nil, err
			}
			servicePlans[i].Spec.ExternalMetadata = &runtime.RawExtension{Raw: metadata}
		}

		if schemas := plan.Schemas; schemas != nil {
			if instanceSchemas := schemas.ServiceInstance; instanceSchemas != nil {
				if instanceCreateSchema := instanceSchemas.Create; instanceCreateSchema != nil && instanceCreateSchema.Parameters != nil {
					schema, err := json.Marshal(instanceCreateSchema.Parameters)
					if err != nil {
						err = fmt.Errorf("Failed to marshal instance create schema \n%+v\n %v", instanceCreateSchema.Parameters, err)
						glog.Error(err)
						return nil, err
					}
					servicePlans[i].Spec.ServiceInstanceCreateParameterSchema = &runtime.RawExtension{Raw: schema}
				}
				if instanceUpdateSchema := instanceSchemas.Update; instanceUpdateSchema != nil && instanceUpdateSchema.Parameters != nil {
					schema, err := json.Marshal(instanceUpdateSchema.Parameters)
					if err != nil {
						err = fmt.Errorf("Failed to marshal instance update schema \n%+v\n %v", instanceUpdateSchema.Parameters, err)
						glog.Error(err)
						return nil, err
					}
					servicePlans[i].Spec.ServiceInstanceUpdateParameterSchema = &runtime.RawExtension{Raw: schema}
				}
			}
			if bindingSchemas := schemas.ServiceBinding; bindingSchemas != nil {
				if bindingCreateSchema := bindingSchemas.Create; bindingCreateSchema != nil {
					if bindingCreateSchema.Parameters != nil {
						schema, err := json.Marshal(bindingCreateSchema.Parameters)
						if err != nil {
							err = fmt.Errorf("Failed to marshal binding create schema \n%+v\n %v", bindingCreateSchema.Parameters, err)
							glog.Error(err)
							return nil, err
						}
						servicePlans[i].Spec.ServiceBindingCreateParameterSchema = &runtime.RawExtension{Raw: schema}
					}
					if utilfeature.DefaultFeatureGate.Enabled(scfeatures.ResponseSchema) && bindingCreateSchema.Response != nil {
						schema, err := json.Marshal(bindingCreateSchema.Response)
						if err != nil {
							err = fmt.Errorf("Failed to marshal binding create response schema \n%+v\n %v", bindingCreateSchema.Response, err)
							glog.Error(err)
							return nil, err
						}
						servicePlans[i].Spec.ServiceBindingCreateResponseSchema = &runtime.RawExtension{Raw: schema}
					}
				}
			}
		}
	}
	return servicePlans, nil
}

// isServiceInstanceConditionTrue returns whether the given instance has a given condition
// with status true.
func isServiceInstanceConditionTrue(instance *v1beta1.ServiceInstance, conditionType v1beta1.ServiceInstanceConditionType) bool {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == conditionType {
			return cond.Status == v1beta1.ConditionTrue
		}
	}

	return false
}

// isServiceInstanceReady returns whether the given instance has a ready condition
// with status true.
func isServiceInstanceReady(instance *v1beta1.ServiceInstance) bool {
	return isServiceInstanceConditionTrue(instance, v1beta1.ServiceInstanceConditionReady)
}

// isServiceInstanceFailed returns whether the instance has a failed condition with
// status true.
func isServiceInstanceFailed(instance *v1beta1.ServiceInstance) bool {
	return isServiceInstanceConditionTrue(instance, v1beta1.ServiceInstanceConditionFailed)
}

// isServiceInstanceOrphanMitigation returns whether the given instance has an
// orphan mitigation condition with status true.
func isServiceInstanceOrphanMitigation(instance *v1beta1.ServiceInstance) bool {
	return isServiceInstanceConditionTrue(instance, v1beta1.ServiceInstanceConditionOrphanMitigation)
}

// NewClientConfigurationForBroker creates a new ClientConfiguration for connecting
// to the specified Broker
func NewClientConfigurationForBroker(meta metav1.ObjectMeta, commonSpec *v1beta1.CommonServiceBrokerSpec, authConfig *osb.AuthConfig) *osb.ClientConfiguration {
	clientConfig := osb.DefaultClientConfiguration()
	clientConfig.Name = meta.Name
	clientConfig.URL = commonSpec.URL
	clientConfig.AuthConfig = authConfig
	clientConfig.EnableAlphaFeatures = true
	clientConfig.Insecure = commonSpec.InsecureSkipTLSVerify
	clientConfig.CAData = commonSpec.CABundle
	return clientConfig
}

// reconciliationRetryDurationExceeded returns whether the given operation
// start time has exceeded the controller's set reconciliation retry duration.
func (c *controller) reconciliationRetryDurationExceeded(operationStartTime *metav1.Time) bool {
	if operationStartTime == nil || time.Now().Before(operationStartTime.Time.Add(c.reconciliationRetryDuration)) {
		return false
	}
	return true
}

// shouldStartOrphanMitigation returns whether an error with the given status
// code indicates that orphan migitation should start.
func shouldStartOrphanMitigation(statusCode int) bool {
	is2XX := statusCode >= 200 && statusCode < 300
	is5XX := statusCode >= 500 && statusCode < 600

	return (is2XX && statusCode != http.StatusOK) || is5XX
}

// isRetriableHTTPStatus returns whether an error with the given HTTP status
// code is retriable.
func isRetriableHTTPStatus(statusCode int) bool {
	return statusCode != http.StatusBadRequest
}

// ReconciliationAction represents a type of action the reconciler should take
// for a resource.
type ReconciliationAction string

const (
	reconcileAdd    ReconciliationAction = "Add"
	reconcileUpdate ReconciliationAction = "Update"
	reconcileDelete ReconciliationAction = "Delete"
	reconcilePoll   ReconciliationAction = "Poll"
)

func (c *controller) getClusterID() (id string) {
	// Use the read lock for reading, so that multiple instances
	// provisioning at the same time do not collide.
	c.clusterIDLock.RLock()
	id = c.clusterID
	c.clusterIDLock.RUnlock()

	// fast exit if it exists
	if id != "" {
		return
	}
	// lazily create on first access if does not exist
	c.clusterIDLock.Lock()
	// check the id again to make sure nobody set ID while we were
	// locking
	if id = c.clusterID; id == "" {
		id = string(uuid.NewUUID())
		c.clusterID = id
	}
	c.clusterIDLock.Unlock()
	return
}

func (c *controller) setClusterID(id string) {
	c.clusterIDLock.Lock()
	c.clusterID = id
	c.clusterIDLock.Unlock()
}

// getServiceClassPlanAndServiceBrokerForServiceBinding is a sequence of operations that's
// done to validate service plan, service class exist, and handles creating
// a brokerclient to use for a given ServiceInstance.
// Sets ServiceClassRef and/or ServicePlanRef if they haven't been already set.
func (c *controller) getServiceClassPlanAndServiceBrokerForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding) (*v1beta1.ServiceClass, *v1beta1.ServicePlan, string, osb.Client, error) {
	serviceClass, serviceBrokerName, osbClient, err := c.getServiceClassAndServiceBrokerForServiceBinding(instance, binding)
	if err != nil {
		return nil, nil, "", nil, err
	}
	servicePlan, err := c.getServicePlanForServiceBinding(instance, binding, serviceClass)
	if err != nil {
		return nil, nil, "", nil, err
	}

	return serviceClass, servicePlan, serviceBrokerName, osbClient, nil
}

func (c *controller) getServiceClassAndServiceBrokerForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding) (*v1beta1.ServiceClass, string, osb.Client, error) {
	serviceClass, err := c.getServiceClassForServiceBinding(instance, binding)
	if err != nil {
		return nil, "", nil, err
	}

	serviceBroker, err := c.getServiceBrokerForServiceBinding(instance, binding, serviceClass)
	if err != nil {
		return nil, "", nil, err
	}

	//osbClient, err := c.getBrokerClientForServiceBinding(instance, binding, serviceBroker)
	osbClient, err := c.getBrokerClientForServiceBinding(instance, binding)
	if err != nil {
		return nil, "", nil, err
	}

	return serviceClass, serviceBroker.Name, osbClient, nil
}

func (c *controller) getServiceClassForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding) (*v1beta1.ServiceClass, error) {
	pcb := pretty.NewInstanceContextBuilder(instance)
	serviceClass, err := c.serviceClassLister.ServiceClasses(instance.Namespace).Get(instance.Spec.ServiceClassRef.Name)
	if err != nil {
		s := fmt.Sprintf(
			"References a non-existent ServiceClass %q - %c",
			instance.Spec.ServiceClassRef.Name, instance.Spec.PlanReference,
		)
		glog.Warning(pcb.Message(s))
		c.updateServiceBindingCondition(
			binding,
			v1beta1.ServiceBindingConditionReady,
			v1beta1.ConditionFalse,
			errorNonexistentClusterServiceClassReason,
			"The binding references a ServiceClass that does not exist. "+s,
		)
		c.recorder.Event(binding, corev1.EventTypeWarning, errorNonexistentClusterServiceClassMessage, s)
		return nil, err
	}
	return serviceClass, nil
}

func (c *controller) getServicePlanForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding, serviceClass *v1beta1.ServiceClass) (*v1beta1.ServicePlan, error) {
	pcb := pretty.NewInstanceContextBuilder(instance)
	servicePlan, err := c.servicePlanLister.ServicePlans(instance.Namespace).Get(instance.Spec.ServicePlanRef.Name)
	if nil != err {
		s := fmt.Sprintf(
			"References a non-existent ServicePlan %q - %v",
			instance.Spec.ServicePlanRef.Name, instance.Spec.PlanReference,
		)
		glog.Warning(pcb.Message(s))
		c.updateServiceBindingCondition(
			binding,
			v1beta1.ServiceBindingConditionReady,
			v1beta1.ConditionFalse,
			errorNonexistentClusterServicePlanReason,
			"The ServiceBinding references an ServiceInstance which references ServicePlan that does not exist. "+s,
		)
		c.recorder.Event(binding, corev1.EventTypeWarning, errorNonexistentClusterServicePlanReason, s)
		return nil, fmt.Errorf(s)
	}
	return servicePlan, nil
}

func (c *controller) getServiceBrokerForServiceBinding(instance *v1beta1.ServiceInstance, binding *v1beta1.ServiceBinding, serviceClass *v1beta1.ServiceClass) (*v1beta1.ServiceBroker, error) {
	pcb := pretty.NewInstanceContextBuilder(instance)

	broker, err := c.serviceBrokerLister.ServiceBrokers(instance.Namespace).Get(serviceClass.Spec.ServiceBrokerName)
	if err != nil {
		s := fmt.Sprintf("References a non-existent ServiceBroker %q", serviceClass.Spec.ServiceBrokerName)
		glog.Warning(pcb.Message(s))
		c.updateServiceBindingCondition(
			binding,
			v1beta1.ServiceBindingConditionReady,
			v1beta1.ConditionFalse,
			errorNonexistentClusterServiceBrokerReason,
			"The binding references a ServiceBroker that does not exist. "+s,
		)
		c.recorder.Event(binding, corev1.EventTypeWarning, errorNonexistentClusterServiceBrokerReason, s)
		return nil, err
	}
	return broker, nil
}
