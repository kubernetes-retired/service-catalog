/*
Copyright 2016 The Kubernetes Authors.

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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	settings "github.com/kubernetes-incubator/service-catalog/pkg/apis/settings/v1alpha1"
	servicecataloginformers "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions"
)

func TestMergeEnv(t *testing.T) {
	tests := map[string]struct {
		orig       []corev1.EnvVar
		mod        []corev1.EnvVar
		result     []corev1.EnvVar
		shouldFail bool
	}{
		"empty original": {
			mod:        []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
			result:     []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
			shouldFail: false,
		},
		"good merge": {
			orig:       []corev1.EnvVar{{Name: "abcd", Value: "value2"}, {Name: "hello", Value: "value3"}},
			mod:        []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
			result:     []corev1.EnvVar{{Name: "abcd", Value: "value2"}, {Name: "hello", Value: "value3"}, {Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
			shouldFail: false,
		},
		"conflict": {
			orig:       []corev1.EnvVar{{Name: "abc", Value: "value3"}},
			mod:        []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
			shouldFail: true,
		},
		"one is exact same": {
			orig:       []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "hello", Value: "value3"}},
			mod:        []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
			result:     []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "hello", Value: "value3"}, {Name: "ABC", Value: "value3"}},
			shouldFail: false,
		},
	}

	for name, test := range tests {
		result, err := mergeEnv(
			test.orig,
			[]*settings.PodPreset{{Spec: settings.PodPresetSpec{Env: test.mod}}},
		)
		if test.shouldFail && err == nil {
			t.Fatalf("expected test %q to fail but got nil", name)
		}
		if !test.shouldFail && err != nil {
			t.Fatalf("test %q failed: %v", name, err)
		}
		if !reflect.DeepEqual(test.result, result) {
			t.Fatalf("results were not equal for test %q: got %#v; expected: %#v", name, result, test.result)
		}
	}
}

func TestMergeEnvFrom(t *testing.T) {
	tests := map[string]struct {
		orig       []corev1.EnvFromSource
		mod        []corev1.EnvFromSource
		result     []corev1.EnvFromSource
		shouldFail bool
	}{
		"empty original": {
			mod: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
				{
					Prefix: "pre_",
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
			},
			result: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
				{
					Prefix: "pre_",
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
			},
			shouldFail: false,
		},
		"good merge": {
			orig: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "thing"},
					},
				},
			},
			mod: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
				{
					Prefix: "pre_",
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
			},
			result: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "thing"},
					},
				},
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
				{
					Prefix: "pre_",
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
			},
			shouldFail: false,
		},
	}

	for name, test := range tests {
		result, err := mergeEnvFrom(
			test.orig,
			[]*settings.PodPreset{{Spec: settings.PodPresetSpec{EnvFrom: test.mod}}},
		)
		if test.shouldFail && err == nil {
			t.Fatalf("expected test %q to fail but got nil", name)
		}
		if !test.shouldFail && err != nil {
			t.Fatalf("test %q failed: %v", name, err)
		}
		if !reflect.DeepEqual(test.result, result) {
			t.Fatalf("results were not equal for test %q: got %#v; expected: %#v", name, result, test.result)
		}
	}
}

func TestMergeVolumeMounts(t *testing.T) {
	tests := map[string]struct {
		orig       []corev1.VolumeMount
		mod        []corev1.VolumeMount
		result     []corev1.VolumeMount
		shouldFail bool
	}{
		"empty original": {
			mod: []corev1.VolumeMount{
				{
					Name:      "simply-mounted-volume",
					MountPath: "/opt/",
				},
			},
			result: []corev1.VolumeMount{
				{
					Name:      "simply-mounted-volume",
					MountPath: "/opt/",
				},
			},
			shouldFail: false,
		},
		"good merge": {
			mod: []corev1.VolumeMount{
				{
					Name:      "simply-mounted-volume",
					MountPath: "/opt/",
				},
			},
			orig: []corev1.VolumeMount{
				{
					Name:      "etc-volume",
					MountPath: "/etc/",
				},
			},
			result: []corev1.VolumeMount{
				{
					Name:      "etc-volume",
					MountPath: "/etc/",
				},
				{
					Name:      "simply-mounted-volume",
					MountPath: "/opt/",
				},
			},
			shouldFail: false,
		},
		"conflict": {
			mod: []corev1.VolumeMount{
				{
					Name:      "simply-mounted-volume",
					MountPath: "/opt/",
				},
				{
					Name:      "etc-volume",
					MountPath: "/things/",
				},
			},
			orig: []corev1.VolumeMount{
				{
					Name:      "etc-volume",
					MountPath: "/etc/",
				},
			},
			shouldFail: true,
		},
		"conflict on mount path": {
			mod: []corev1.VolumeMount{
				{
					Name:      "simply-mounted-volume",
					MountPath: "/opt/",
				},
				{
					Name:      "things-volume",
					MountPath: "/etc/",
				},
			},
			orig: []corev1.VolumeMount{
				{
					Name:      "etc-volume",
					MountPath: "/etc/",
				},
			},
			shouldFail: true,
		},
		"one is exact same": {
			mod: []corev1.VolumeMount{
				{
					Name:      "simply-mounted-volume",
					MountPath: "/opt/",
				},
				{
					Name:      "etc-volume",
					MountPath: "/etc/",
				},
			},
			orig: []corev1.VolumeMount{
				{
					Name:      "etc-volume",
					MountPath: "/etc/",
				},
			},
			result: []corev1.VolumeMount{
				{
					Name:      "etc-volume",
					MountPath: "/etc/",
				},
				{
					Name:      "simply-mounted-volume",
					MountPath: "/opt/",
				},
			},
			shouldFail: false,
		},
	}

	for name, test := range tests {
		result, err := mergeVolumeMounts(
			test.orig,
			[]*settings.PodPreset{{Spec: settings.PodPresetSpec{VolumeMounts: test.mod}}},
		)
		if test.shouldFail && err == nil {
			t.Fatalf("expected test %q to fail but got nil", name)
		}
		if !test.shouldFail && err != nil {
			t.Fatalf("test %q failed: %v", name, err)
		}
		if !reflect.DeepEqual(test.result, result) {
			t.Fatalf("results were not equal for test %q: got %#v; expected: %#v", name, result, test.result)
		}
	}
}

func TestMergeVolumes(t *testing.T) {
	tests := map[string]struct {
		orig       []corev1.Volume
		mod        []corev1.Volume
		result     []corev1.Volume
		shouldFail bool
	}{
		"empty original": {
			mod: []corev1.Volume{
				{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol2", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			result: []corev1.Volume{
				{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol2", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			shouldFail: false,
		},
		"good merge": {
			orig: []corev1.Volume{
				{Name: "vol3", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol4", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			mod: []corev1.Volume{
				{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol2", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			result: []corev1.Volume{
				{Name: "vol3", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol4", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol2", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			shouldFail: false,
		},
		"conflict": {
			orig: []corev1.Volume{
				{Name: "vol3", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol4", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			mod: []corev1.Volume{
				{Name: "vol3", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/etc/apparmor.d"}}},
				{Name: "vol2", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			shouldFail: true,
		},
		"one is exact same": {
			orig: []corev1.Volume{
				{Name: "vol3", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol4", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			mod: []corev1.Volume{
				{Name: "vol3", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol2", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			result: []corev1.Volume{
				{Name: "vol3", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol4", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol2", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			shouldFail: false,
		},
	}

	for name, test := range tests {
		result, err := mergeVolumes(
			test.orig,
			[]*settings.PodPreset{{Spec: settings.PodPresetSpec{Volumes: test.mod}}},
		)
		if test.shouldFail && err == nil {
			t.Fatalf("expected test %q to fail but got nil", name)
		}
		if !test.shouldFail && err != nil {
			t.Fatalf("test %q failed: %v", name, err)
		}
		if !reflect.DeepEqual(test.result, result) {
			t.Fatalf("results were not equal for test %q: got %#v; expected: %#v", name, result, test.result)
		}
	}
}

// NewTestAdmission provides an admission plugin with test implementations of internal structs.  It uses
// an authorizer that always returns true.
// func NewTestAdmission(lister settingslisters.PodPresetLister, objects ...runtime.Object) kadmission.Interface {
// Build a test client that the admission plugin can use to look up the service account missing from its cache
// 	client := fake.NewSimpleClientset(objects...)
//
// 	return &podPresetPlugin{
// 		client:  client,
// 		Handler: kadmission.NewHandler(kadmission.Create),
// 		lister:  lister,
// 	}
// }

func TestAdmitConflictWithDifferentNamespaceShouldDoNothing(t *testing.T) {
	containerName := "container"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "namespace",
			Labels: map[string]string{
				"security": "S2",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: containerName,
					Env:  []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
				},
			},
		},
	}

	pip := &settings.PodPreset{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hello",
			Namespace: "othernamespace",
		},
		Spec: settings.PodPresetSpec{
			Selector: v1.LabelSelector{
				MatchExpressions: []v1.LabelSelectorRequirement{
					{
						Key:      "security",
						Operator: v1.LabelSelectorOpIn,
						Values:   []string{"S2"},
					},
				},
			},
			Env: []corev1.EnvVar{{Name: "abc", Value: "value"}, {Name: "ABC", Value: "value"}},
		},
	}

	err := admitPod(pod, pip)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdmitConflictWithNonMatchingLabelsShouldNotError(t *testing.T) {
	containerName := "container"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "namespace",
			Labels: map[string]string{
				"security": "S2",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: containerName,
					Env:  []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
				},
			},
		},
	}

	pip := &settings.PodPreset{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hello",
			Namespace: "namespace",
		},
		Spec: settings.PodPresetSpec{
			Selector: v1.LabelSelector{
				MatchExpressions: []v1.LabelSelectorRequirement{
					{
						Key:      "security",
						Operator: v1.LabelSelectorOpIn,
						Values:   []string{"S3"},
					},
				},
			},
			Env: []corev1.EnvVar{{Name: "abc", Value: "value"}, {Name: "ABC", Value: "value"}},
		},
	}

	err := admitPod(pod, pip)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdmitConflictShouldNotModifyPod(t *testing.T) {
	containerName := "container"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "namespace",
			Labels: map[string]string{
				"security": "S2",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: containerName,
					Env:  []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABC", Value: "value3"}},
				},
			},
		},
	}
	origPod := *pod

	pip := &settings.PodPreset{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hello",
			Namespace: "namespace",
		},
		Spec: settings.PodPresetSpec{
			Selector: v1.LabelSelector{
				MatchExpressions: []v1.LabelSelectorRequirement{
					{
						Key:      "security",
						Operator: v1.LabelSelectorOpIn,
						Values:   []string{"S2"},
					},
				},
			},
			Env: []corev1.EnvVar{{Name: "abc", Value: "value"}, {Name: "ABC", Value: "value"}},
		},
	}

	err := admitPod(pod, pip)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&origPod, pod) {
		t.Fatalf("pod should not get modified in case of conflict origPod: %+v got: %+v", &origPod, pod)
	}
}

func TestAdmit(t *testing.T) {
	containerName := "container"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "namespace",
			Labels: map[string]string{
				"security": "S2",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: containerName,
					Env:  []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABCD", Value: "value3"}},
				},
			},
		},
	}

	pip := &settings.PodPreset{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hello",
			Namespace: "namespace",
		},
		Spec: settings.PodPresetSpec{
			Selector: v1.LabelSelector{
				MatchExpressions: []v1.LabelSelectorRequirement{
					{
						Key:      "security",
						Operator: v1.LabelSelectorOpIn,
						Values:   []string{"S2"},
					},
				},
			},
			Volumes: []corev1.Volume{{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			Env:     []corev1.EnvVar{{Name: "abcd", Value: "value"}, {Name: "ABC", Value: "value"}},
			EnvFrom: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
				{
					Prefix: "pre_",
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
			},
		},
	}

	err := admitPod(pod, pip)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdmitMirrorPod(t *testing.T) {
	containerName := "container"

	mirrorPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "namespace",
			Labels: map[string]string{
				"security": "S2",
			},
			Annotations: map[string]string{corev1.MirrorPodAnnotationKey: "mirror"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: containerName,
				},
			},
		},
	}

	pip := &settings.PodPreset{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hello",
			Namespace: "namespace",
		},
		Spec: settings.PodPresetSpec{
			Selector: v1.LabelSelector{
				MatchExpressions: []v1.LabelSelectorRequirement{
					{
						Key:      "security",
						Operator: v1.LabelSelectorOpIn,
						Values:   []string{"S2"},
					},
				},
			},
			Volumes: []corev1.Volume{{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			Env:     []corev1.EnvVar{{Name: "abcd", Value: "value"}, {Name: "ABC", Value: "value"}},
			EnvFrom: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
				{
					Prefix: "pre_",
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
			},
		},
	}

	if err := admitPod(mirrorPod, pip); err != nil {
		t.Fatal(err)
	}

	container := mirrorPod.Spec.Containers[0]
	if len(mirrorPod.Spec.Volumes) != 0 ||
		len(container.VolumeMounts) != 0 ||
		len(container.Env) != 0 ||
		len(container.EnvFrom) != 0 {
		t.Fatalf("mirror pod is updated by PodPreset admission:\n\tVolumes got %d, expected 0\n\tVolumeMounts go %d, expected 0\n\tEnv got, %d expected 0\n\tEnvFrom got %d, expected 0", len(mirrorPod.Spec.Volumes), len(container.VolumeMounts), len(container.Env), len(container.EnvFrom))
	}
}

func TestExclusionNoAdmit(t *testing.T) {
	containerName := "container"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "namespace",
			Labels: map[string]string{
				"security": "S2",
			},
			Annotations: map[string]string{
				corev1.PodPresetOptOutAnnotationKey: "true",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: containerName,
					Env:  []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABCD", Value: "value3"}},
				},
			},
		},
	}

	pip := &settings.PodPreset{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hello",
			Namespace: "namespace",
		},
		Spec: settings.PodPresetSpec{
			Selector: v1.LabelSelector{
				MatchExpressions: []v1.LabelSelectorRequirement{
					{
						Key:      "security",
						Operator: v1.LabelSelectorOpIn,
						Values:   []string{"S2"},
					},
				},
			},
			Volumes: []corev1.Volume{{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			Env:     []corev1.EnvVar{{Name: "abcd", Value: "value"}, {Name: "ABC", Value: "value"}},
			EnvFrom: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
				{
					Prefix: "pre_",
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
			},
		},
	}
	// originalPod, err := corev1.Scheme.Copy(pod)
	originalPod, err := runtime.NewScheme().DeepCopy(pod)
	if err != nil {
		t.Fatal(err)
	}

	err = admitPod(pod, pip)
	if err != nil {
		t.Fatal(err)
	}

	// verify PodSpec has not been mutated
	if !reflect.DeepEqual(pod, originalPod) {
		t.Fatalf("Expected pod spec of '%v' to be unchanged", pod.Name)
	}
}

func TestAdmitEmptyPodNamespace(t *testing.T) {
	containerName := "container"

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mypod",
			Labels: map[string]string{
				"security": "S2",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: containerName,
					Env:  []corev1.EnvVar{{Name: "abc", Value: "value2"}, {Name: "ABCD", Value: "value3"}},
				},
			},
		},
	}

	pip := &settings.PodPreset{
		ObjectMeta: v1.ObjectMeta{
			Name:      "hello",
			Namespace: "different", // (pod will be submitted to namespace 'namespace')
		},
		Spec: settings.PodPresetSpec{
			Selector: v1.LabelSelector{
				MatchExpressions: []v1.LabelSelectorRequirement{
					{
						Key:      "security",
						Operator: v1.LabelSelectorOpIn,
						Values:   []string{"S2"},
					},
				},
			},
			Volumes: []corev1.Volume{{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			Env:     []corev1.EnvVar{{Name: "abcd", Value: "value"}, {Name: "ABC", Value: "value"}},
			EnvFrom: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
				{
					Prefix: "pre_",
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "abc"},
					},
				},
			},
		},
	}
	// originalPod, err := corev1.Scheme.Copy(pod)
	originalPod, err := runtime.NewScheme().DeepCopy(pod)
	if err != nil {
		t.Fatal(err)
	}

	err = admitPod(pod, pip)
	if err != nil {
		t.Fatal(err)
	}

	// verify PodSpec has not been mutated
	if !reflect.DeepEqual(pod, originalPod) {
		t.Fatalf("pod should not get modified in case of emptyNamespace origPod: %+v got: %+v", originalPod, pod)
	}
}

func admitPod(pod *corev1.Pod, pip *settings.PodPreset) error {
	recorder := record.NewFakeRecorder(10)
	informerFactory := servicecataloginformers.NewSharedInformerFactory(
		nil,
		0,
	)
	// All shared informers are v1alpha1 API level
	store := informerFactory.Settings().V1alpha1().PodPresets().Informer().GetStore()
	store.Add(pip)

	return admit(pod, informerFactory.Settings().V1alpha1().PodPresets().Lister(), recorder)
}
