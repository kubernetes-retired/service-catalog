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

// Package app implements a server that runs the service catalog controllers.
package app

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"k8s.io/client-go/1.5/kubernetes"
	v1core "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/1.5/tools/record"

	"k8s.io/kubernetes/pkg/client/typed/discovery"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/healthz"
	"k8s.io/kubernetes/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/util/configz"
	"k8s.io/kubernetes/pkg/util/wait"

	// The API groups for our API must be installed before we can use the
	// client to work with them.  This needs to be done once per process; this
	// is the point at which we handle this for the controller-manager
	// process.  Please do not remove.
	_ "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/install"
	// The core API has to be installed in order for the client to understand
	// error messages from the API server.  Please do not remove.
	_ "k8s.io/kubernetes/pkg/api/install"

	"github.com/kubernetes-incubator/service-catalog/cmd/controller-manager/app/options"
	"github.com/kubernetes-incubator/service-catalog/pkg/brokerapi/openservicebroker"
	servicecataloginformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// NewControllerManagerCommand creates a *cobra.Command object with default
// parameters.
func NewControllerManagerCommand() *cobra.Command {
	s := options.NewControllerManagerServer()
	s.AddFlags(pflag.CommandLine)
	cmd := &cobra.Command{
		Use: "controller-manager",
		Long: `The service-catalog controller manager is a daemon that embeds
the core control loops shipped with the service catalog.`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	return cmd
}

const controllerManagerAgentName = "service-catalog-controller-manager"
const controllerDiscoveryAgentName = "service-catalog-controller-discovery"

// Run runs the service-catalog controller-manager; should never exit.
func Run(controllerManagerOptions *options.ControllerManagerServer) error {
	// TODO: what does this do

	// if c, err := configz.New("componentconfig"); err == nil {
	// 	c.Set(controllerManagerOptions.KubeControllerManagerConfiguration)
	// } else {
	// 	glog.Errorf("unable to register configz: %s", err)
	// }

	// Build the K8s kubeconfig / client / clientBuilder
	glog.V(4).Info("Building k8s kubeconfig")

	k8sKubeconfig, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Failed to get kube client config (%s)", err)
	}
	k8sKubeconfig.GroupVersion = &unversioned.GroupVersion{}

	// k8sKubeconfig, err := clientcmd.BuildConfigFromFlags(
	// 	controllerManagerOptions.K8sAPIServerURL,
	// 	controllerManagerOptions.K8sKubeconfigPath)
	// if err != nil {
	// 	// TODO: disambiguate API errors
	// 	return err
	// }
	k8sKubeconfig.ContentConfig.ContentType = controllerManagerOptions.ContentType
	// Override kubeconfig qps/burst settings from flags
	k8sKubeconfig.QPS = controllerManagerOptions.KubeAPIQPS
	k8sKubeconfig.Burst = int(controllerManagerOptions.KubeAPIBurst)
	k8sKubeClient, err := kubernetes.NewForConfig(
		rest.AddUserAgent(k8sKubeconfig, controllerManagerAgentName),
	)
	if err != nil {
		glog.Fatalf("Invalid k8s API configuration: %v", err)
	}

	glog.V(4).Infof("Building service-catalog kubeconfig for url: %v\n", controllerManagerOptions.ServiceCatalogAPIServerURL)
	// Build the service-catalog kubeconfig / clientBuilder
	serviceCatalogKubeconfig, err := clientcmd.BuildConfigFromFlags(
		controllerManagerOptions.ServiceCatalogAPIServerURL,
		// controllerManagerOptions.ServiceCatalogKubeconfigPath,
		"", // TODO: tolerate missing kubeconfig
	)
	if err != nil {
		// TODO: disambiguate API errors
		return err
	}

	glog.V(4).Info("Starting http server and mux")
	// Start http server and handlers
	go func() {
		mux := http.NewServeMux()
		healthz.InstallHandler(mux)
		configz.InstallHandler(mux)

		server := &http.Server{
			Addr:    net.JoinHostPort(controllerManagerOptions.Address, strconv.Itoa(int(controllerManagerOptions.Port))),
			Handler: mux,
		}
		glog.Fatal(server.ListenAndServe())
	}()

	// Create event broadcaster
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: k8sKubeClient.Core().Events("")})
	recorder := eventBroadcaster.NewRecorder(v1.EventSource{Component: controllerManagerAgentName})

	// 'run' is the logic to run the controllers for the controller manager
	run := func(stop <-chan struct{}) {
		serviceCatalogClientBuilder := controller.SimpleClientBuilder{
			ClientConfig: serviceCatalogKubeconfig,
		}

		// TODO: understand service account story for this controller-manager

		// if len(s.ServiceAccountKeyFile) > 0 && controllerManagerOptions.UseServiceAccountCredentials {
		// 	k8sClientBuilder = controller.SAControllerClientBuilder{
		// 		ClientConfig: restclient.AnonymousClientConfig(k8sKubeconfig),
		// 		CoreClient:   k8sKubeClient.Core(),
		// 		Namespace:    "kube-system",
		// 	}
		// } else {
		// 	k8sClientBuilder = rootClientBuilder
		// }

		err := StartControllers(controllerManagerOptions, k8sKubeconfig, serviceCatalogClientBuilder, recorder, stop)
		glog.Fatalf("error running controllers: %v", err)
		panic("unreachable")
	}

	// TODO: leader election for this controller-manager
	// id, err := os.Hostname()
	// if err != nil {
	// 	return err
	// }
	// rl := resourcelock.EndpointsLock{
	// 	EndpointsMeta: v1.ObjectMeta{
	// 		Namespace: "kube-system",
	// 		Name:      "kube-controller-manager",
	// 	},
	// 	Client: leaderElectionClient,
	// 	LockConfig: resourcelock.ResourceLockConfig{
	// 		Identity:      id,
	// 		EventRecorder: recorder,
	// 	},
	// }

	// leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
	// 	Lock:          &rl,
	// 	LeaseDuration: controllerManagerOptions.LeaderElection.LeaseDuration.Duration,
	// 	RenewDeadline: controllerManagerOptions.LeaderElection.RenewDeadline.Duration,
	// 	RetryPeriod:   controllerManagerOptions.LeaderElection.RetryPeriod.Duration,
	// 	Callbacks: leaderelection.LeaderCallbacks{
	// 		OnStartedLeading: run,
	// 		OnStoppedLeading: func() {
	// 			glog.Fatalf("leaderelection lost")
	// 		},
	// 	},
	// })
	// panic("unreachable")

	run(make(<-chan (struct{})))

	return nil
}

// getAvailableResources uses the discovery client to determine which API
// groups are available in the endpoint reachable from the given client and
// returns a map of them.
func getAvailableResources(clientBuilder controller.ClientBuilder) (map[schema.GroupVersionResource]bool, error) {
	var discoveryClient discovery.DiscoveryInterface

	// If apiserver is not running we should wait for some time and fail only then. This is particularly
	// important when we start apiserver and controller manager at the same time.
	err := wait.PollImmediate(time.Second, 10*time.Second, func() (bool, error) {
		client, err := clientBuilder.Client(controllerDiscoveryAgentName)
		if err != nil {
			glog.Errorf("Failed to get api versions from server: %v", err)
			return false, nil
		}

		glog.V(4).Info("Created client for discovery")

		discoveryClient = client.Discovery()
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get api versions from server: %v", err)
	}

	resourceMap, err := discoveryClient.ServerResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get supported resources from server: %v", err)
	}

	allResources := map[schema.GroupVersionResource]bool{}
	for _, apiResourceList := range resourceMap {
		glog.V(4).Infof("Resource: %#v", apiResourceList)
		version, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			return nil, err
		}
		for _, apiResource := range apiResourceList.APIResources {
			allResources[version.WithResource(apiResource.Name)] = true
		}
	}

	return allResources, nil
}

// StartControllers starts all the controllers in the service-catalog
// controller manager.
func StartControllers(s *options.ControllerManagerServer,
	coreKubeconfig *rest.Config,
	serviceCatalogClientBuilder controller.ClientBuilder,
	recorder record.EventRecorder,
	stop <-chan struct{}) error {

	// Get available service-catalog resources
	glog.V(5).Info("Getting available resources")
	availableResources, err := getAvailableResources(serviceCatalogClientBuilder)
	if err != nil {
		return err
	}

	resyncDuration := 1 * time.Minute

	coreKubeconfig = rest.AddUserAgent(coreKubeconfig, controllerManagerAgentName)
	coreClient, err := kubernetes.NewForConfig(coreKubeconfig)
	if err != nil {
		glog.Fatal(err)
	}

	// Launch service-catalog controller
	if availableResources[schema.GroupVersionResource{Group: "servicecatalog.k8s.io", Version: "v1alpha1", Resource: "brokers"}] {
		glog.V(5).Info("Creating shared informers")
		// Build the informer factory for service-catalog resources
		informerFactory := servicecataloginformers.NewSharedInformerFactory(
			nil, // internal clientset (not used)
			serviceCatalogClientBuilder.ClientOrDie("shared-informers"),
			resyncDuration,
		)
		// All shared informers are v1alpha1 API level
		serviceCatalogSharedInformers := informerFactory.Servicecatalog().V1alpha1()

		glog.V(5).Info("Creating controller")
		serviceCatalogController, err := controller.NewController(
			coreClient,
			serviceCatalogClientBuilder.ClientOrDie(controllerManagerAgentName).ServicecatalogV1alpha1(),
			serviceCatalogSharedInformers.Brokers(),
			serviceCatalogSharedInformers.ServiceClasses(),
			serviceCatalogSharedInformers.Instances(),
			serviceCatalogSharedInformers.Bindings(),
			openservicebroker.NewClient,
		)
		if err != nil {
			return err
		}

		glog.V(5).Info("Running controller")
		go serviceCatalogController.Run(stop)

		glog.V(1).Info("Starting shared informers")
		informerFactory.Start(stop)
	} else {
		glog.V(1).Infof("Skipping starting service-catalog controller because servicecatalog/v1alpha1 is not available")
	}

	select {}
}
