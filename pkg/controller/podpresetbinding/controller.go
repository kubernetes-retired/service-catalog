package podpresetbinding

import (
	"log"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	settingsv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings/v1alpha1"
	settingsv1alpha1client "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset/versioned/typed/settings/v1alpha1"
	settingsv1alpha1informer "github.com/kubernetes-incubator/service-catalog/pkg/client/informers/externalversions/settings/v1alpha1"
	settingsv1alpha1lister "github.com/kubernetes-incubator/service-catalog/pkg/client/listers/settings/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/inject/args"
	servicecatalogclient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
)

// EDIT THIS FILE
// This files was created by "kubebuilder create resource" for you to edit.
// Controller implementation logic for PodPresetBinding resources goes here.

func getCatalogClient() *servicecatalogclient.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}
	clientset, err := servicecatalogclient.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}

	return clientset
}

func (bc *PodPresetBindingController) Reconcile(k types.ReconcileKey) error {
	// INSERT YOUR CODE HERE
	log.Printf("Implement the Reconcile function on podpresetbinding.PodPresetBindingController to reconcile %s\n", k.Name)

	clientset := getCatalogClient()

	return nil
}

// +kubebuilder:controller:group=settings,version=v1alpha1,kind=PodPresetBinding,resource=podpresetbindings
type PodPresetBindingController struct {
	// INSERT ADDITIONAL FIELDS HERE
	podpresetbindingLister settingsv1alpha1lister.PodPresetBindingLister
	podpresetbindingclient settingsv1alpha1client.SettingsV1alpha1Interface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	podpresetbindingrecorder record.EventRecorder
}

// ProvideController provides a controller that will be run at startup.  Kubebuilder will use codegeneration
// to automatically register this controller in the inject package
func ProvideController(arguments args.InjectArgs) (*controller.GenericController, error) {
	// INSERT INITIALIZATIONS FOR ADDITIONAL FIELDS HERE
	bc := &PodPresetBindingController{
		podpresetbindingLister:   arguments.ControllerManager.GetInformerProvider(&settingsv1alpha1.PodPresetBinding{}).(settingsv1alpha1informer.PodPresetBindingInformer).Lister(),
		podpresetbindingclient:   arguments.Clientset.SettingsV1alpha1(),
		podpresetbindingrecorder: arguments.CreateRecorder("PodPresetBindingController"),
	}

	// Create a new controller that will call PodPresetBindingController.Reconcile on changes to PodPresetBindings
	gc := &controller.GenericController{
		Name:             "PodPresetBindingController",
		Reconcile:        bc.Reconcile,
		InformerRegistry: arguments.ControllerManager,
	}
	if err := gc.Watch(&settingsv1alpha1.PodPresetBinding{}); err != nil {
		return gc, err
	}

	// IMPORTANT:
	// To watch additional resource types - such as those created by your controller - add gc.Watch* function calls here
	// Watch function calls will transform each object event into a PodPresetBinding Key to be reconciled by the controller.
	//
	// **********
	// For any new Watched types, you MUST add the appropriate // +kubebuilder:informer and // +kubebuilder:rbac
	// annotations to the PodPresetBindingController and run "kubebuilder generate.
	// This will generate the code to start the informers and create the RBAC rules needed for running in a cluster.
	// See:
	// https://godoc.org/github.com/kubernetes-sigs/kubebuilder/pkg/gen/controller#example-package
	// **********

	return gc, nil
}
