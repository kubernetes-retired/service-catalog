package podpreset

import (
	"log"

	"github.com/kubernetes-sigs/kubebuilder/pkg/controller"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	settingsv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings/v1alpha1"
	settingsv1alpha1client "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset/versioned/typed/settings/v1alpha1"
	settingsv1alpha1informer "github.com/kubernetes-incubator/service-catalog/pkg/client/informers/externalversions/settings/v1alpha1"
	settingsv1alpha1lister "github.com/kubernetes-incubator/service-catalog/pkg/client/listers/settings/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/inject/args"
)

// EDIT THIS FILE
// This files was created by "kubebuilder create resource" for you to edit.
// Controller implementation logic for PodPreset resources goes here.

func (bc *PodPresetController) Reconcile(k types.ReconcileKey) error {
	// INSERT YOUR CODE HERE
	log.Printf("Implement the Reconcile function on podpreset.PodPresetController to reconcile %s\n", k.Name)

	return nil
}

// +kubebuilder:controller:group=settings,version=v1alpha1,kind=PodPreset,resource=podpresets
type PodPresetController struct {
	// INSERT ADDITIONAL FIELDS HERE
	podpresetLister settingsv1alpha1lister.PodPresetLister
	podpresetclient settingsv1alpha1client.SettingsV1alpha1Interface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	podpresetrecorder record.EventRecorder
}

// ProvideController provides a controller that will be run at startup.  Kubebuilder will use codegeneration
// to automatically register this controller in the inject package
func ProvideController(arguments args.InjectArgs) (*controller.GenericController, error) {
	// INSERT INITIALIZATIONS FOR ADDITIONAL FIELDS HERE
	bc := &PodPresetController{
		podpresetLister:   arguments.ControllerManager.GetInformerProvider(&settingsv1alpha1.PodPreset{}).(settingsv1alpha1informer.PodPresetInformer).Lister(),
		podpresetclient:   arguments.Clientset.SettingsV1alpha1(),
		podpresetrecorder: arguments.CreateRecorder("PodPresetController"),
	}

	// Create a new controller that will call PodPresetController.Reconcile on changes to PodPresets
	gc := &controller.GenericController{
		Name:             "PodPresetController",
		Reconcile:        bc.Reconcile,
		InformerRegistry: arguments.ControllerManager,
	}
	if err := gc.Watch(&settingsv1alpha1.PodPreset{}); err != nil {
		return gc, err
	}

	// IMPORTANT:
	// To watch additional resource types - such as those created by your controller - add gc.Watch* function calls here
	// Watch function calls will transform each object event into a PodPreset Key to be reconciled by the controller.
	//
	// **********
	// For any new Watched types, you MUST add the appropriate // +kubebuilder:informer and // +kubebuilder:rbac
	// annotations to the PodPresetController and run "kubebuilder generate.
	// This will generate the code to start the informers and create the RBAC rules needed for running in a cluster.
	// See:
	// https://godoc.org/github.com/kubernetes-sigs/kubebuilder/pkg/gen/controller#example-package
	// **********
	if err := gc.Watch(&v1.Pod{}); err != nil {
		return gc, err
	}

	return gc, nil
}
