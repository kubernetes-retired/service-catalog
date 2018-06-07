package inject

import (
	settingsv1alpha1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings/v1alpha1"
	rscheme "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset/versioned/scheme"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/podpreset"
	"github.com/kubernetes-incubator/service-catalog/pkg/inject/args"
	"github.com/kubernetes-sigs/kubebuilder/pkg/inject/run"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	rscheme.AddToScheme(scheme.Scheme)

	// Inject Informers
	Inject = append(Inject, func(arguments args.InjectArgs) error {
		Injector.ControllerManager = arguments.ControllerManager

		if err := arguments.ControllerManager.AddInformerProvider(&settingsv1alpha1.PodPreset{}, arguments.Informers.Settings().V1alpha1().PodPresets()); err != nil {
			return err
		}
		if err := arguments.ControllerManager.AddInformerProvider(&settingsv1alpha1.PodPresetBinding{}, arguments.Informers.Settings().V1alpha1().PodPresetBindings()); err != nil {
			return err
		}

		// Add Kubernetes informers
		if err := arguments.ControllerManager.AddInformerProvider(&corev1.Pod{}, arguments.KubernetesInformers.Core().V1().Pods()); err != nil {
			return err
		}

		if c, err := podpreset.ProvideController(arguments); err != nil {
			return err
		} else {
			arguments.ControllerManager.AddController(c)
		}
		if c, err := podpresetbinding.ProvideController(arguments); err != nil {
			return err
		} else {
			arguments.ControllerManager.AddController(c)
		}
		return nil
	})

	// Inject CRDs
	Injector.CRDs = append(Injector.CRDs, &settingsv1alpha1.PodPresetCRD)
	Injector.CRDs = append(Injector.CRDs, &settingsv1alpha1.PodPresetBindingCRD)
	// Inject PolicyRules
	Injector.PolicyRules = append(Injector.PolicyRules, rbacv1.PolicyRule{
		APIGroups: []string{"settings.servicecatalog.k8s.io"},
		Resources: []string{"*"},
		Verbs:     []string{"*"},
	})
	Injector.PolicyRules = append(Injector.PolicyRules, rbacv1.PolicyRule{
		APIGroups: []string{
			"",
		},
		Resources: []string{
			"pods",
		},
		Verbs: []string{
			"get", "list", "watch",
		},
	})
	// Inject GroupVersions
	Injector.GroupVersions = append(Injector.GroupVersions, schema.GroupVersion{
		Group:   "settings.servicecatalog.k8s.io",
		Version: "v1alpha1",
	})
	Injector.RunFns = append(Injector.RunFns, func(arguments run.RunArguments) error {
		Injector.ControllerManager.RunInformersAndControllers(arguments)
		return nil
	})
}
