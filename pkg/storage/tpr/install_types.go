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

package tpr

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	extensionsv1beta "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5/typed/extensions/v1beta1"
	"k8s.io/kubernetes/pkg/util/wait"
)

// this is the set of third party resources to be installed. each key is the name of the TPR to
// install, and each value is the resource to install
//
var thirdPartyResources = []v1beta1.ThirdPartyResource{
	serviceBrokerTPR,
	serviceClassTPR,
	serviceInstanceTPR,
	serviceBindingTPR,
}

// Installer takes in a client and exposes method for installing third party resources
type Installer struct {
	tprs extensionsv1beta.ThirdPartyResourceInterface
}

// NewInstaller is used to install third party resources
func NewInstaller(tprs extensionsv1beta.ThirdPartyResourceInterface) *Installer {
	return &Installer{
		tprs: tprs,
	}
}

// InstallTypes installs all third party resource types to the cluster
func (i *Installer) InstallTypes() error {
	var wg sync.WaitGroup
	errMsg := make(chan string, len(thirdPartyResources))

	for _, tpr := range thirdPartyResources {
		glog.Infof("Checking for existence of %s", tpr.Name)
		if _, err := i.tprs.Get(tpr.Name); err == nil {
			glog.Infof("Found existing TPR %s", tpr.Name)
			continue
		}

		glog.Infof("Creating Third Party Resource Type: %s", tpr.Name)

		wg.Add(1)
		go func(tpr v1beta1.ThirdPartyResource, client extensionsv1beta.ThirdPartyResourceInterface) {
			defer wg.Done()
			if _, err := i.tprs.Create(&tpr); err != nil {
				errMsg <- fmt.Sprintf("%s: %s", tpr.Name, err)
			} else {
				glog.Infof("Created TPR '%s'", tpr.Name)

				// There can be a delay, so poll until it's ready to go...
				err := wait.PollImmediate(1*time.Second, 1*time.Second, func() (bool, error) {
					if _, err := client.Get(tpr.Name); err == nil {
						glog.Infof("TPR %s is ready", tpr.Name)
						return true, nil
					}

					glog.Infof("TPR %s is not ready yet... waiting...", tpr.Name)
					return false, nil
				})
				if err != nil {
					glog.Infof("Error polling for TPR status:", err)
				}
			}
		}(tpr, i.tprs)
	}

	wg.Wait()
	close(errMsg)

	var allErrMsg string
	for msg := range errMsg {
		if msg != "" {
			allErrMsg = fmt.Sprintf("%s\n%s", allErrMsg, msg)
		}
	}

	if allErrMsg != "" {
		glog.Errorf("Failed to create Third Party Resource:\n%s)", allErrMsg)
		return fmt.Errorf("Failed to create Third Party Resource:\n%s)", allErrMsg)
	}

	return nil
}
