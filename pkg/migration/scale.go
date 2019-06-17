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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

// ScalingService can be used to scale up or down configured deployment
type ScalingService struct {
	namespace      string
	deploymentName string
	appInterface   appsv1.AppsV1Interface
}

// NewScalingService creates a new ScalingService instance.
func NewScalingService(namespace string, deploymentName string, appInterface appsv1.AppsV1Interface) *ScalingService {
	return &ScalingService{
		namespace:      namespace,
		deploymentName: deploymentName,
		appInterface:   appInterface,
	}
}

// ScaleDown scales down the deployment
func (s *ScalingService) ScaleDown() error {
	klog.Infoln("Scaling down the controller")
	return s.scaleTo(0)
}

// ScaleUp scales up the deployment
func (s *ScalingService) ScaleUp() error {
	klog.Infoln("Scaling up the controller")
	return s.scaleTo(1)
}

func (s *ScalingService) scaleTo(v int) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		deploy, err := s.appInterface.Deployments(s.namespace).Get(s.deploymentName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		r := int32(v)
		deploy.Spec.Replicas = &r
		_, err = s.appInterface.Deployments(s.namespace).Update(deploy)
		return err
	})
	if err != nil {
		return err
	}

	err = wait.Poll(time.Second, time.Second*45, func() (bool, error) {
		deploy, err := s.appInterface.Deployments(s.namespace).Get(s.deploymentName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if deploy.Status.ReadyReplicas == *deploy.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})

	return err
}
