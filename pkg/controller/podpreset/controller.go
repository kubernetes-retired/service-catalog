package podpreset

import (
	"fmt"
	"log"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	settingsv1alpha1 "github.com/jpeeler/podpreset-crd/pkg/apis/settings/v1alpha1"
	settingsv1alpha1client "github.com/jpeeler/podpreset-crd/pkg/client/clientset/versioned/typed/settings/v1alpha1"
	settingsv1alpha1informer "github.com/jpeeler/podpreset-crd/pkg/client/informers/externalversions/settings/v1alpha1"
	settingsv1alpha1lister "github.com/jpeeler/podpreset-crd/pkg/client/listers/settings/v1alpha1"
	"github.com/jpeeler/podpreset-crd/pkg/inject/args"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	annotationPrefix = "podpreset.admission.kubernetes.io"
)

// EDIT THIS FILE
// This files was created by "kubebuilder create resource" for you to edit.
// Controller implementation logic for PodPreset resources goes here.

// Get a clientset with in-cluster config.
// TODO: share with webhook's config.go?
func getClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}

	return clientset
}

func (bc *PodPresetController) Reconcile(k types.ReconcileKey) error {
	// INSERT YOUR CODE HERE
	log.Printf("Implement the Reconcile function on podpreset.PodPresetController to reconcile %s\n", k.Name)

	pp, err := bc.podpresetclient.PodPresets(k.Namespace).Get(k.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	clientset := getClient()
	selector, _ := metav1.LabelSelectorAsSelector(&pp.Spec.Selector)
	deploymentList, err := clientset.AppsV1().Deployments(k.Namespace).List(metav1.ListOptions{})

	for i, deployment := range deploymentList.Items {
		glog.V(6).Infof("(%v) Looking at deployment %v\n", i, deployment.Name)
		if selector.Matches(labels.Set(deployment.Spec.Template.ObjectMeta.Labels)) {
			bouncedKey := fmt.Sprintf("%s/bounced-%s", annotationPrefix, pp.GetName())
			resourceVersion, found := deployment.Spec.Template.ObjectMeta.Annotations[bouncedKey]
			if !found || found && resourceVersion < pp.GetResourceVersion() {
				// bounce pod since this is the first mutation or a later mutation has occurred
				glog.V(4).Infof("Detected deployment '%v' needs bouncing", deployment.Name)
				metav1.SetMetaDataAnnotation(&deployment.Spec.Template.ObjectMeta, bouncedKey, pp.GetResourceVersion())
				_, err = clientset.AppsV1().Deployments(k.Namespace).Update(&deployment)
				if err != nil {
					return err
				}
			}
		}
	}

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
	// if err := gc.Watch(&v1.Pod{}); err != nil {
	// 	return gc, err
	// }

	return gc, nil
}
