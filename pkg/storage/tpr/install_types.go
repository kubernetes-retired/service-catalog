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
	"time"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
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

// InstallTypes installs all third party resource types to the cluster
func InstallTypes(cs clientset.Interface) error {
	tprs := cs.Extensions().ThirdPartyResources()
	for _, tpr := range thirdPartyResources {
		glog.Infof("Checking for existence of %s", tpr.Name)
		if _, err := tprs.Get(tpr.Name); err == nil {
			glog.Infof("Found existing TPR %s", tpr.Name)
			continue
		}

		glog.Infof("Creating Third Party Resource Type: %s", tpr.Name)
		if _, err := tprs.Create(&tpr); err != nil {
			glog.Errorf("Failed to create Third Party Resource Type: %s (%s))", tpr.Name, err)
			return err
		}
		glog.Infof("Created TPR '%s'", tpr.Name)
		// There can be a delay, so poll until it's ready to go...
		for i := 0; i < 30; i++ {
			if _, err := tprs.Get(tpr.Name); err != nil {
				glog.Infof("TPR %s is ready", tpr.Name)
				break
			}
			glog.Infof("TPR %s is not ready yet... waiting...", tpr.Name)
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}
