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

package framework

import (
	goflag "flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/golang/glog"
	v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	util "github.com/kubernetes-incubator/service-catalog/test/util"
	"github.com/spf13/cobra"
	pflag "github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var options *HealthCheckServer

// Execute starts the HTTP Server and runs the health check tasks on a periodic basis
func Execute() error {
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	options = NewHealthCheckServer()
	options.AddFlags(pflag.CommandLine)
	pflag.CommandLine.Set("alsologtostderr", "true")
	defer glog.Flush()
	return rootCmd.Execute()

}

var rootCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "healthcheck performs an end to end verification of Service Catalog",
	Long: "healthcheck monitors the health of Service Catalog and exposes Prometheus " +
		"metrics for centralized monitoring and alerting.  Once started, " +
		"healthcheck runs tasks on a periodic basis that verifies end to end " +
		"Service Catalog functionality. This testing requires a Service Broker (such " +
		"as the UPS Broker or OSB Stub broker) is deployed.  Both of these brokers are designed " +
		"for testing and do not actually create or manage any services.",
	Run: func(cmd *cobra.Command, args []string) {
		h, err := NewHealthCheck(options)
		if err != nil {
			glog.Errorf("Error initialzing: %v", err)
			os.Exit(1)
		}

		// Start the HTTP server that enables us to serve /healtz and /metrics.   The  metrics can be pulled,
		// analyzed and alerted on.
		err = ServeHTTP(options)
		if err != nil {
			glog.Errorf("Error starting HTTP: %v", err)
			os.Exit(1)
		}

		glog.Infof("Scheduled health checks will be run every %v", options.HealthCheckInterval)

		// Every X interval run the health check
		ticker := time.NewTicker(options.HealthCheckInterval)
		for range ticker.C {
			h.RunHealthCheck(options)
		}
	},
}

// HealthCheck is a type that used to control various aspects of the health
// check.
type HealthCheck struct {
	kubeClientSet           kubernetes.Interface
	serviceCatalogClientSet clientset.Interface
	brokername              string
	brokernamespace         string
	serviceclassName        string
	serviceclassID          string
	serviceplanID           string
	instanceName            string
	bindingName             string
	brokerendpointName      string
	namespace               *corev1.Namespace
	frameworkError          error
}

// NewHealthCheck creates a new HealthCheck object and initializes the kube
// and catalog client sets.
func NewHealthCheck(s *HealthCheckServer) (*HealthCheck, error) {
	h := &HealthCheck{}
	var kubeConfig *rest.Config

	err := h.initBrokerAttributes(s)
	if err != nil {
		return nil, err
	}

	// If token exists assume we are running in a pod
	_, err = os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err == nil {
		kubeConfig, err = rest.InClusterConfig()
	} else {
		kubeConfig, err = LoadConfig(s.KubeConfig, s.KubeContext)
	}

	if err != nil {
		return nil, err
	}

	h.kubeClientSet, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		glog.Errorf("Error creating kubeClientSet: %v", err)
		return nil, err
	}

	h.serviceCatalogClientSet, err = clientset.NewForConfig(kubeConfig)
	if err != nil {
		glog.Errorf("Error creating serviceCatalogClientSet: %v", err)
		return nil, err
	}

	return h, nil
}

// RunHealthCheck runs an end to end verification against the "ups-broker".  It
// validates the broker endpoint is available, then creates an instance and
// binding and does validation along the way and then tears it down.  Some basic
// Prometheus metrics are maintained that can be alerted off from.
func (h *HealthCheck) RunHealthCheck(s *HealthCheckServer) error {
	ExecutionCount.Inc()
	hcStartTime := time.Now()

	h.verifyBrokerIsReady()
	h.createNamespace()
	h.createInstance()
	h.createBinding()
	h.deprovision()
	h.deleteNamespace()

	if h.frameworkError == nil {
		ReportOperationCompleted("healthcheck_completed", hcStartTime)
		glog.V(2).Infof("Successfully ran health check in %v", time.Since(hcStartTime))
		glog.V(4).Info("") // for readabilty/separation of test runs
	} else {
		h.cleanup()
		ErrorCount.WithLabelValues(h.frameworkError.Error()).Inc()
	}
	return h.frameworkError
}

// verifyBrokerIsReady verifies the Broker is found and appears ready
func (h *HealthCheck) verifyBrokerIsReady() error {
	h.frameworkError = nil
	glog.V(4).Infof("checking for endpoint %v/%v", h.brokernamespace, h.brokerendpointName)
	err := WaitForEndpoint(h.kubeClientSet, h.brokernamespace, h.brokerendpointName)
	if err != nil {
		return h.setError("endpoint not found: %v", err.Error())
	}

	url := "http://" + h.brokername + "." + h.brokernamespace + ".svc.cluster.local"
	broker := &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: h.brokername,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}

	glog.V(4).Infof("checking for Broker %v to be ready", broker.Name)
	err = util.WaitForBrokerCondition(h.serviceCatalogClientSet.ServicecatalogV1beta1(),
		broker.Name,
		v1beta1.ServiceBrokerCondition{
			Type:   v1beta1.ServiceBrokerConditionReady,
			Status: v1beta1.ConditionTrue,
		},
	)
	if err != nil {
		return h.setError("broker not ready: %v", err.Error())
	}

	err = util.WaitForClusterServiceClassToExist(h.serviceCatalogClientSet.ServicecatalogV1beta1(), h.serviceclassID)
	if err != nil {
		return h.setError("service class not found: %v", err.Error())
	}
	return nil
}

// createInstance creates a Service Instance and verifies it becomes ready
// and it's references are resolved
func (h *HealthCheck) createInstance() error {
	if h.frameworkError != nil {
		return h.frameworkError
	}
	glog.V(4).Info("Creating a ServiceInstance")
	instance := &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.instanceName,
			Namespace: h.namespace.Name,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: h.serviceclassName,
				ClusterServicePlanExternalName:  "default",
			},
		},
	}
	operationStartTime := time.Now()
	var err error
	instance, err = h.serviceCatalogClientSet.ServicecatalogV1beta1().ServiceInstances(h.namespace.Name).Create(instance)
	if err != nil {
		return h.setError("error creating instance: %v", err.Error())
	}

	if instance == nil {
		return h.setError("error creating instance - instance is null", "")
	}

	glog.V(4).Info("Waiting for ServiceInstance to be ready")
	err = util.WaitForInstanceCondition(h.serviceCatalogClientSet.ServicecatalogV1beta1(),
		h.namespace.Name,
		h.instanceName,
		v1beta1.ServiceInstanceCondition{
			Type:   v1beta1.ServiceInstanceConditionReady,
			Status: v1beta1.ConditionTrue,
		},
	)
	if err != nil {
		return h.setError("instance not ready: %v", err.Error())
	}
	ReportOperationCompleted("create_instance", operationStartTime)

	glog.V(4).Info("Verifing references are resolved")
	sc, err := h.serviceCatalogClientSet.ServicecatalogV1beta1().ServiceInstances(h.namespace.Name).Get(h.instanceName, metav1.GetOptions{})
	if err != nil {
		return h.setError("error getting instance: %v", err.Error())
	}

	if sc.Spec.ClusterServiceClassRef == nil {
		return h.setError("ClusterServiceClassRef should not be null", "")
	}
	if sc.Spec.ClusterServicePlanRef == nil {
		return h.setError("ClusterServicePlanRef should not be null", "")
	}

	if strings.Compare(sc.Spec.ClusterServiceClassRef.Name, h.serviceclassID) != 0 {
		return h.setError("ClusterServiceClassRef.Name should not be null", "")
	}
	if strings.Compare(sc.Spec.ClusterServicePlanRef.Name, h.serviceplanID) != 0 {
		return h.setError("ClusterServicePlanRef.Name should not be null", "")
	}
	return nil
}

// createBinding creates a binding and verifies the binding and secret are
// correct
func (h *HealthCheck) createBinding() error {
	if h.frameworkError != nil {
		return h.frameworkError
	}
	glog.V(4).Info("Creating a ServiceBinding")
	binding := &v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.bindingName,
			Namespace: h.namespace.Name,
		},
		Spec: v1beta1.ServiceBindingSpec{
			ServiceInstanceRef: v1beta1.LocalObjectReference{
				Name: h.instanceName,
			},
			SecretName: "my-secret",
		},
	}
	operationStartTime := time.Now()
	binding, err := h.serviceCatalogClientSet.ServicecatalogV1beta1().ServiceBindings(h.namespace.Name).Create(binding)
	if err != nil {
		return h.setError("Error creating binding: %v", err.Error())
	}
	if binding == nil {
		return h.setError("Binding should not be null", "")
	}

	glog.V(4).Info("Waiting for ServiceBinding to be ready")
	_, err = util.WaitForBindingCondition(h.serviceCatalogClientSet.ServicecatalogV1beta1(),
		h.namespace.Name,
		h.bindingName,
		v1beta1.ServiceBindingCondition{
			Type:   v1beta1.ServiceBindingConditionReady,
			Status: v1beta1.ConditionTrue,
		},
	)
	if err != nil {
		return h.setError("binding not ready: %v", err.Error())
	}
	ReportOperationCompleted("binding_ready", operationStartTime)

	glog.V(4).Info("Validating that a secret was created after binding")
	_, err = h.kubeClientSet.CoreV1().Secrets(h.namespace.Name).Get("my-secret", metav1.GetOptions{})
	if err != nil {
		return h.setError("Error getting secret: %v", err.Error())
	}
	glog.V(4).Info("Successfully created instance & binding.  Cleaning up.")
	return nil
}

// deprovision deletes the service binding, deprovisions the service instance
// and verifies it does the appropriate cleanup.
func (h *HealthCheck) deprovision() error {
	if h.frameworkError != nil {
		return h.frameworkError
	}
	glog.V(4).Info("Deleting the ServiceBinding.")
	operationStartTime := time.Now()
	err := h.serviceCatalogClientSet.ServicecatalogV1beta1().ServiceBindings(h.namespace.Name).Delete(h.bindingName, nil)
	if err != nil {
		return h.setError("error deleting binding: %v", err.Error())
	}

	glog.V(4).Info("Waiting for ServiceBinding to be removed")
	err = util.WaitForBindingToNotExist(h.serviceCatalogClientSet.ServicecatalogV1beta1(), h.namespace.Name, h.bindingName)
	if err != nil {
		return h.setError("binding not removed: %v", err.Error())
	}
	ReportOperationCompleted("binding_deleted", operationStartTime)

	glog.V(4).Info("Verifying that the secret was deleted after deleting the binding")
	_, err = h.kubeClientSet.CoreV1().Secrets(h.namespace.Name).Get("my-secret", metav1.GetOptions{})
	if err == nil {
		return h.setError("secret not deleted", "")
	}

	// Deprovisioning the ServiceInstance
	glog.V(4).Info("Deleting the ServiceInstance")
	operationStartTime = time.Now()
	err = h.serviceCatalogClientSet.ServicecatalogV1beta1().ServiceInstances(h.namespace.Name).Delete(h.instanceName, nil)
	if err != nil {
		return h.setError("error deleting instance: %v", err.Error())
	}

	glog.V(4).Info("Waiting for ServiceInstance to be removed")
	err = util.WaitForInstanceToNotExist(h.serviceCatalogClientSet.ServicecatalogV1beta1(), h.namespace.Name, h.instanceName)
	if err != nil {
		return h.setError("instance not removed: %v", err.Error())
	}
	ReportOperationCompleted("instance_deleted", operationStartTime)
	return nil
}

// cleanup is invoked when the healthcheck test fails.  It should delete any residue from the test.
func (h *HealthCheck) cleanup() {
	if h.namespace != nil {
		glog.V(4).Infof("Cleaning up.  Deleting the binding, instance and test namespace %v", h.namespace.Name)
		h.serviceCatalogClientSet.ServicecatalogV1beta1().ServiceBindings(h.namespace.Name).Delete(h.bindingName, nil)
		h.serviceCatalogClientSet.ServicecatalogV1beta1().ServiceInstances(h.namespace.Name).Delete(h.instanceName, nil)
		DeleteKubeNamespace(h.kubeClientSet, h.namespace.Name)
		h.namespace = nil
	}
}

func (h *HealthCheck) createNamespace() error {
	if h.frameworkError != nil {
		return h.frameworkError
	}
	var err error
	h.namespace, err = CreateKubeNamespace(h.kubeClientSet)
	if err != nil {
		h.setError(err.Error(), "%v")
	}
	return nil
}

func (h *HealthCheck) deleteNamespace() error {
	if h.frameworkError != nil {
		return h.frameworkError
	}
	err := DeleteKubeNamespace(h.kubeClientSet, h.namespace.Name)
	if err != nil {
		return h.setError("failed to delete namespace: %v", err.Error())
	}
	h.namespace = nil
	return err
}

func (h *HealthCheck) initBrokerAttributes(s *HealthCheckServer) error {
	switch s.TestBrokerName {
	case "ups-broker":
		h.brokername = "ups-broker"
		h.brokernamespace = "ups-broker"
		h.brokerendpointName = "ups-broker-ups-broker"
		h.serviceclassName = "user-provided-service"
		h.serviceclassID = "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468"
		h.serviceplanID = "86064792-7ea2-467b-af93-ac9694d96d52"
		h.instanceName = "ups-instance"
		h.bindingName = "ups-binding"
	case "osb-stub":
		h.brokername = "osb-stub"
		h.brokernamespace = "osb-stub"
		h.brokerendpointName = "osb-stub"
		h.serviceclassName = "noop-service"
		h.serviceclassID = "0861dc50-beed-4f9d-ba97-e78f43b802da"
		h.serviceplanID = "977715c5-4a12-452f-994a-4caf4f8cba02"
		h.instanceName = "stub-instance"
		h.bindingName = "stub-binding"
	default:
		return fmt.Errorf("invalid broker-name specified: %v.  Valid options are ups-broker and stub-broker", s.TestBrokerName)
	}
	return nil
}

// setError creates a new error using msg and param for the formated message.
// The message is logged and the HealthCheck error state is set and returned.
// This function attempts to log the location of the caller (file name & line
// number) so as to maintain context of where the error occured
func (h *HealthCheck) setError(msg, param string) error {
	_, file, line, _ := runtime.Caller(1)

	// only use the last 30 characters
	context := len(file) - 30
	if context < 0 {
		context = 0
	}
	partialFileName := file[context:]
	format := fmt.Sprintf("...%s:%d: %v", partialFileName, line, msg)
	h.frameworkError = fmt.Errorf(format, param)
	glog.Info(h.frameworkError.Error())
	return h.frameworkError
}
