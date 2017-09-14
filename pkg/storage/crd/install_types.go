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

package crd

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// this is the set of custome resources to be installed. each key is the name of the CRD to
// install, and each value is the resource to install
//
var customResourceDefinitions = []*crdv1beta1.CustomResourceDefinition{
	serviceBrokerCRD,
	serviceClassCRD,
	serviceInstanceCRD,
	serviceInstanceCredentialCRD,
}

// ErrCRDInstall is returned when we fail to install CRD
type ErrCRDInstall struct {
	errMsg string
}

func (e ErrCRDInstall) Error() string {
	return e.errMsg
}

// InstallTypes installs all third party resource types to the cluster
func InstallTypes(cl crdclient.CustomResourceDefinitionInterface) error {
	var wg sync.WaitGroup
	errMsg := make(chan string, len(customResourceDefinitions))

	for _, crd := range customResourceDefinitions {
		glog.Infof("Checking for existence of %s", crd.Name)
		if _, err := cl.Get(crd.Name, metav1.GetOptions{}); err == nil {
			glog.Infof("Found existing CRD %s", crd.Name)
			continue
		}

		glog.Infof("Creating Custom Resource Type: %s", crd.Name)

		wg.Add(1)
		go func(crd crdv1beta1.CustomResourceDefinition, client crdclient.CustomResourceDefinitionInterface) {
			defer wg.Done()
			if _, err := cl.Create(&crd); err != nil {
				errMsg <- fmt.Sprintf("%s: %s", crd.Name, err)
			} else {
				glog.Infof("Created CRD '%s'", crd.Name)

				// There can be a delay, so poll until it's ready to go...
				err := wait.PollImmediate(1*time.Second, 1*time.Second, func() (bool, error) {
					if _, err := client.Get(crd.Name, metav1.GetOptions{}); err == nil {
						glog.Infof("CRD %s is ready", crd.Name)
						return true, nil
					}

					glog.Infof("CRD %s is not ready yet... waiting...", crd.Name)
					return false, nil
				})
				if err != nil {
					glog.Infof("Error polling for CRD status:", err)
				}
			}
		}(*crd, cl)
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
		glog.Errorf("Failed to create Custom Resource:\n%s", allErrMsg)
		return ErrCRDInstall{fmt.Sprintf("Failed to create Custom Resource: %s", allErrMsg)}
	}

	return nil
}
