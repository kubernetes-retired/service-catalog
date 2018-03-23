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
	"errors"
	goflag "flag"
	"fmt"
	"os"
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

func Execute() error {
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	options = NewHealthCheckServer()
	options.AddFlags(pflag.CommandLine)
	defer glog.Flush()
	return rootCmd.Execute()

}

var rootCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "healtchcheck performs an end to end verification of Service Catalog",
	Long:  `Runs a quick end to end health check for Service Catalog`,
	Run: func(cmd *cobra.Command, args []string) {
		err := initialize(options)
		if err != nil {
			glog.Errorf("Error initialzing: %v", err)
			os.Exit(1)
		}

		// Start the HTTP server that enables us to serve /healtz and /metrics.   The  metrics can be pulled,
		// analyzed and alerted on.
		err = ServeHttp(options)
		if err != nil {
			glog.Errorf("Error starting HTTP: %v", err)
			os.Exit(1)
		}

		glog.Infof("Sheduled health checks will be run every %v", options.HealthCheckInterval)

		// Every X interval run the health check
		ticker := time.NewTicker(options.HealthCheckInterval)
		for range ticker.C {
			err = healthCheck(options)
			if err != nil {
				glog.Errorf("Error running health check: %v", err)
				cleanup()
				ErrorCount.WithLabelValues(err.Error()).Inc()
			} else {
				glog.V(2).Info("Successfully ran health check")
			}
		}
	},
}

var (
	// A Kubernetes and Service Catalog client
	kubeClientSet           kubernetes.Interface
	serviceCatalogClientSet clientset.Interface

	// Namespace in which all test resources should reside
	namespace        *corev1.Namespace
	upsbrokername    = "ups-broker"
	serviceclassName = "user-provided-service"
	serviceclassID   = "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468"
	serviceplanID    = "86064792-7ea2-467b-af93-ac9694d96d52"
	instanceName     = "ups-instance"
	bindingName      = "ups-binding"
)

func initialize(s *HealthCheckServer) error {
	var kubeConfig *rest.Config

	// If token exists assume we are running in a pod
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err == nil {
		kubeConfig, err = rest.InClusterConfig()
	} else {
		kubeConfig, err = LoadConfig(s.KubeConfig, s.KubeContext)
	}

	if err != nil {
		return err
	}

	kubeClientSet, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		glog.Errorf("Error creating kubeClientSet: %v", err)
		return err
	}

	serviceCatalogClientSet, err = clientset.NewForConfig(kubeConfig)
	if err != nil {
		glog.Errorf("Error creating serviceCatalogClientSet: %v", err)
		return err
	}

	return nil
}

// healthCheck runs an end to end verification against the "ups-broker".  It
// validates the broker endpoint is available, then creates an instance and
// binding and does validation along the way and then tears it down.  Some basic
// Prometheus metrics are maintained that can be alerted off from.
func healthCheck(s *HealthCheckServer) error {
	var e error
	ExecutionCount.Inc()
	hcStartTime := time.Now()
	glog.V(4).Infof("checking for %v", upsbrokername)
	err := WaitForEndpoint(kubeClientSet, "ups-broker", "ups-broker-ups-broker")
	if err != nil {
		e = fmt.Errorf("no broker endpoint: %v", err)
		glog.Error(e)
		return e
	}

	url := "http://" + upsbrokername + "." + "ups-broker" + ".svc.cluster.local"
	broker := &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: upsbrokername,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}

	err = util.WaitForBrokerCondition(serviceCatalogClientSet.ServicecatalogV1beta1(),
		broker.Name,
		v1beta1.ServiceBrokerCondition{
			Type:   v1beta1.ServiceBrokerConditionReady,
			Status: v1beta1.ConditionTrue,
		},
	)
	if err != nil {
		e = fmt.Errorf("broker not ready: %v", err)
		glog.Error(e)
		return e
	}

	err = util.WaitForClusterServiceClassToExist(serviceCatalogClientSet.ServicecatalogV1beta1(), serviceclassID)
	if err != nil {
		e = fmt.Errorf("service class not found: %v", err)
		glog.Error(e)
		return e
	}

	namespace, err = CreateKubeNamespace(kubeClientSet)
	if err != nil {
		e = fmt.Errorf("error creating namespace: %v", err)
		glog.Error(e)
		return e
	}

	glog.V(4).Info("Creating a ServiceInstance")
	instance := &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: namespace.Name,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: serviceclassName,
				ClusterServicePlanExternalName:  "default",
			},
		},
	}
	operationStartTime := time.Now()
	instance, err = serviceCatalogClientSet.ServicecatalogV1beta1().ServiceInstances(namespace.Name).Create(instance)
	if err != nil {
		e = fmt.Errorf("error creating instance: %v", err)
		glog.Error(e)
		return e
	}

	if instance == nil {
		e = fmt.Errorf("error creating instance - instance is null")
		glog.Error(e)
		return e
	}

	glog.V(4).Info("Waiting for ServiceInstance to be ready")
	err = util.WaitForInstanceCondition(serviceCatalogClientSet.ServicecatalogV1beta1(),
		namespace.Name,
		instanceName,
		v1beta1.ServiceInstanceCondition{
			Type:   v1beta1.ServiceInstanceConditionReady,
			Status: v1beta1.ConditionTrue,
		},
	)
	if err != nil {
		e = fmt.Errorf("instance not ready: %v", err)
		glog.Error(e)
		return e
	}
	ReportOperationCompleted("create_instance", operationStartTime)

	glog.V(4).Info("Verify references are resolved")
	sc, err := serviceCatalogClientSet.ServicecatalogV1beta1().ServiceInstances(namespace.Name).Get(instanceName, metav1.GetOptions{})
	if err != nil {
		e = fmt.Errorf("error getting instance: %v", err)
		return e
	}

	if sc.Spec.ClusterServiceClassRef == nil {
		err = errors.New("ClusterServiceClassRef should not be null")
		glog.Error(err)
		return err
	}
	if sc.Spec.ClusterServicePlanRef == nil {
		err = errors.New("ClusterServicePlanRef should not be null")
		glog.Error(err)
		return err
	}

	if strings.Compare(sc.Spec.ClusterServiceClassRef.Name, serviceclassID) != 0 {
		err = errors.New("ClusterServiceClassReName should not be null")
		glog.Error(err)
		return err
	}
	if strings.Compare(sc.Spec.ClusterServicePlanRef.Name, serviceplanID) != 0 {
		err = errors.New("ClusterServicePlanReName should not be null")
		glog.Error(err)
		return err
	}

	glog.V(4).Info("Creating a ServiceBinding")
	binding := &v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace.Name,
		},
		Spec: v1beta1.ServiceBindingSpec{
			ServiceInstanceRef: v1beta1.LocalObjectReference{
				Name: instanceName,
			},
			SecretName: "my-secret",
		},
	}
	operationStartTime = time.Now()
	binding, err = serviceCatalogClientSet.ServicecatalogV1beta1().ServiceBindings(namespace.Name).Create(binding)
	if err != nil {
		err = fmt.Errorf("Error creating binding: %v", err)
		glog.Error(err)
		return err
	}
	if binding == nil {
		err = fmt.Errorf("Binding should not be null")
		glog.Error(err)
		return err
	}

	glog.V(4).Info("Waiting for ServiceBinding to be ready")
	_, err = util.WaitForBindingCondition(serviceCatalogClientSet.ServicecatalogV1beta1(),
		namespace.Name,
		bindingName,
		v1beta1.ServiceBindingCondition{
			Type:   v1beta1.ServiceBindingConditionReady,
			Status: v1beta1.ConditionTrue,
		},
	)
	if err != nil {
		e = fmt.Errorf("binding not ready: %v", err)
		glog.Error(e)
		return e
	}
	ReportOperationCompleted("binding_ready", operationStartTime)

	glog.V(4).Info("Validating that a secret was created after binding")
	_, err = kubeClientSet.CoreV1().Secrets(namespace.Name).Get("my-secret", metav1.GetOptions{})
	if err != nil {
		e = fmt.Errorf("Error getting secret: %v", err)
		glog.Error(e)
		return e
	}
	glog.V(4).Info("Successfully created instance & binding.  Cleaning up.")

	glog.V(4).Info("Deleting the ServiceBinding.")
	operationStartTime = time.Now()
	err = serviceCatalogClientSet.ServicecatalogV1beta1().ServiceBindings(namespace.Name).Delete(bindingName, nil)
	if err != nil {
		e = fmt.Errorf("error deleting binding: %v", err)
		glog.Error(e)
		return e
	}

	glog.V(4).Info("Waiting for ServiceBinding to be removed")
	err = util.WaitForBindingToNotExist(serviceCatalogClientSet.ServicecatalogV1beta1(), namespace.Name, bindingName)
	if err != nil {
		e = fmt.Errorf("binding not removed: %v", err)
		glog.Error(e)
		return e
	}
	ReportOperationCompleted("binding_deleted", operationStartTime)

	glog.V(4).Info("Verifying that the secret was deleted after deleting the binding")
	_, err = kubeClientSet.CoreV1().Secrets(namespace.Name).Get("my-secret", metav1.GetOptions{})
	if err == nil {
		e = errors.New("secret not deleted")
		glog.Error(e)
		return e
	}

	// Deprovisioning the ServiceInstance
	glog.V(4).Info("Deleting the ServiceInstance")
	operationStartTime = time.Now()
	err = serviceCatalogClientSet.ServicecatalogV1beta1().ServiceInstances(namespace.Name).Delete(instanceName, nil)
	if err != nil {
		e = fmt.Errorf("error deleting instance: %v", err)
		glog.Error(e)
		return e
	}

	glog.V(4).Info("Waiting for ServiceInstance to be removed")
	err = util.WaitForInstanceToNotExist(serviceCatalogClientSet.ServicecatalogV1beta1(), namespace.Name, instanceName)
	if err != nil {
		e = fmt.Errorf("instance not removed: %v", err)
		glog.Error(e)
		return e
	}
	ReportOperationCompleted("instance_deleted", operationStartTime)

	glog.V(4).Info("Deleting the test namespace")
	err = DeleteKubeNamespace(kubeClientSet, namespace.Name)
	if err != nil {
		e = fmt.Errorf("error deleting namespace: %v", err)
		glog.Error(e)
		return e
	}

	ReportOperationCompleted("healthcheck_completed", hcStartTime)
	return nil
}

// cleanup is invoked when the healthcheck test fails.  It should delete any residue from the test.
// We rely on deletion of the namespace to remove any leftover objects
func cleanup() {
	if namespace != nil {
		glog.V(4).Infof("Cleaning up.  Deleting the test namespace %v", namespace.Name)

		// only a binding should block an instance from being deleted, ensure the binding
		// has been deleted
		serviceCatalogClientSet.ServicecatalogV1beta1().ServiceBindings(namespace.Name).Delete(bindingName, nil)

		err := DeleteKubeNamespace(kubeClientSet, namespace.Name)
		if err != nil {
			glog.V(4).Infof("Failed to delete namespace: %v", err)
		}
		namespace = nil
	}
}
