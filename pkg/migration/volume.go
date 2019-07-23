/*
Copyright 2019 The Kubernetes Authors.

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

package migration

import (
	"github.com/pkg/errors"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// AssertPersistentVolumeClaimDeleted deletes PVC resource in which backup data will be kept and make sure it was removed
func (m *Service) AssertPersistentVolumeClaimDeleted(name string) error {
	klog.Info("Deleting PersistentVolumeClaim")
	err := m.coreInterface.PersistentVolumeClaims(m.releaseNamespace).Delete(name, &metav1.DeleteOptions{})
	if apiErr.IsNotFound(err) {
		klog.Info("PersistentVolumeClaim was removed")
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "while deleting PersistentVolumeClaim")
	}

	// PVC will be removed after erase "kubernetes.io/pvc-protection" finalizer
	// https://kubernetes.io/docs/concepts/storage/persistent-volumes/#storage-object-in-use-protection
	return nil
}
