package podpresetbinding

import (
	"fmt"
	"log"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	settingsv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings/v1alpha1"
	crdversioned "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset/versioned"
	settingsv1alpha1client "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset/versioned/typed/settings/v1alpha1"
	servicecatalogclient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	settingsv1alpha1informer "github.com/kubernetes-incubator/service-catalog/pkg/client/informers/externalversions/settings/v1alpha1"
	settingsv1alpha1lister "github.com/kubernetes-incubator/service-catalog/pkg/client/listers/settings/v1alpha1"
	"github.com/kubernetes-incubator/service-catalog/pkg/inject/args"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func getCrdClient() *crdversioned.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}
	clientset, err := crdversioned.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}

	return clientset
}

// Reconcile is the loop that implements the continuous periodic action logic for this resource
func (bc *PodPresetBindingController) Reconcile(k types.ReconcileKey) error {
	// INSERT YOUR CODE HERE
	log.Printf("Implement the Reconcile function on podpresetbinding.PodPresetBindingController to reconcile %s\n", k.Name)

	ppb, err := bc.podpresetbindingclient.PodPresetBindings(k.Namespace).Get(k.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	clientset := getCatalogClient()
	binding, err := clientset.Servicecatalog().ServiceBindings(k.Namespace).Get(ppb.Spec.BindingRef.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			glog.V(6).Info("Service binding not yet created")
			return nil
		}
		return err
	}

	if len(binding.Status.Conditions) > 0 && binding.Status.Conditions[len(binding.Status.Conditions)-1].Type == v1beta1.ServiceBindingConditionReady {
		// create pod preset if binding status is ready
		crdClientset := getCrdClient()
		if _, err := crdClientset.SettingsV1alpha1().PodPresets(k.Namespace).Create(&ppb.Spec.PodPresetTemplate); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return fmt.Errorf("Unable to create podpreset: %v", err)
			}
			// don't think this is needed
			// if _, err := crdClientset.SettingsV1alpha1().PodPresets(k.Namespace).Update(&ppb.Spec.PodPresetTemplate); err != nil {
			// 	return fmt.Errorf("Unable to update podpreset: %v", err)
			// }
		}
	}

	return nil
}

// PodPresetBindingController is made up of items that are necessary for the reconcile loop to function
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
