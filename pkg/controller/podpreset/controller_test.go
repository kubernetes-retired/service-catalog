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

package podpreset

import (
	"testing"

	servicecataloginformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"

	settingsapi "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings/v1alpha1"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers"
	clientgofake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

var alwaysReady = func() bool { return true }

func TestPodInitializeWithConflict(t *testing.T) {
	// create a fake kube client
	kc := &clientgofake.Clientset{}
	// create a fake sc client
	sc := &servicecatalogclientset.Clientset{}

	coreInformers := coreinformers.NewSharedInformerFactory(kc, 0)
	podInformer := coreInformers.Core().V1().Pods().Informer()
	// create informers
	scInformers := servicecataloginformers.NewSharedInformerFactory(sc, 0)
	podpresetInformer := scInformers.Settings().V1alpha1().PodPresets()

	fakeRecorder := record.NewFakeRecorder(5)

	ir, err := NewController(kc, fakeRecorder, podInformer, podpresetInformer)
	if err != nil {
		t.Errorf("error creating initializer: %v", err)
	}

	podName := "pod1"
	ns := "test-ns"
	podPresetName := "podpreset1"

	pod := newPod(podName, ns)
	podInformer.GetStore().Add(pod)
	podpresetInformer.Informer().GetStore().Add(newPodPreset(podPresetName, ns))

	err = ir.initFn(ns + "/" + podName)
	if err != nil {
		t.Errorf("error initializing pod: %v", err)
	}

	kcActions := kc.Actions()

	// verify pod update has been called to initialize the Pod
	if kcActions[0].GetVerb() != "update" {
		t.Errorf("expected Pod update action, instead of got: %v", kcActions[0].GetVerb())
	}

	events := getRecordedEvents(fakeRecorder)
	if len(events) != 1 { /* conflict event should be published */
		t.Errorf("expected an events, but got %d events", len(events))
	}
}

// newPod returns an instance of test Pod
func newPod(name, namespace string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"security": "S2",
			},
			Initializers: &metav1.Initializers{
				Pending: []metav1.Initializer{
					{Name: podPresetInitializerName},
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: name,
					Env: []corev1.EnvVar{
						{Name: "abc", Value: "value2"},
						{Name: "ABC", Value: "value3"},
					},
				},
			},
		},
	}

}

func newPodPreset(name, namespace string) *settingsapi.PodPreset {
	return &settingsapi.PodPreset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: settingsapi.PodPresetSpec{
			Selector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "security",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"S2"},
					},
				},
			},
			Env: []corev1.EnvVar{{Name: "abc", Value: "value"}, {Name: "ABC", Value: "value"}},
		},
	}
}

func getRecordedEvents(recorder *record.FakeRecorder) []string {
	source := recorder.Events
	done := false
	events := []string{}
	for !done {
		select {
		case event := <-source:
			events = append(events, event)
		default:
			done = true
		}
	}
	return events
}
