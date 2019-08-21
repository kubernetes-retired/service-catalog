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

package readiness

import (
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	waitInterval    = 5 * time.Second
	timeoutInterval = 120 * time.Second
)

// ClientGetter is an interface to represent structs return kubernetes clientset
type ClientGetter interface {
	KubernetesClient() kubernetes.Interface
}

type readiness struct {
	client kubernetes.Interface
	cfg    ServiceCatalogConfig
}

// NewReadiness returns pointer to rediness probe
func NewReadiness(c ClientGetter, scConfig ServiceCatalogConfig) *readiness {
	return &readiness{
		client: c.KubernetesClient(),
		cfg:    scConfig,
	}
}

// TestEnvironmentIsReady runs probe to check all required pods are running
func (r *readiness) TestEnvironmentIsReady() error {
	klog.Info("Assert all pods required to test are ready")
	for _, fn := range []func() error{
		r.assertServiceCatalogIsReady,
		r.assertTestBrokerIsReady,
	} {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *readiness) assertServiceCatalogIsReady() error {
	klog.Info("Make sure ServiceCatalog ApiServer is up")
	if err := r.assertServiceCatalogApiServerIsUp(); err != nil {
		return errors.Wrap(err, "failed during waiting for ServiceCatalog ApiServer")
	}
	klog.Info("ServiceCatalog ApiServer is ready")

	klog.Info("Make sure ServiceCatalog Controller is up")
	if err := r.assertServiceCatalogControllerIsUp(); err != nil {
		return errors.Wrap(err, "failed during waiting for ServiceCatalog Controller")
	}
	klog.Info("ServiceCatalog Controller is ready")

	return nil
}

func (r *readiness) assertServiceCatalogApiServerIsUp() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		deployment, err := r.client.AppsV1beta1().Deployments(r.cfg.ServiceCatalogNamespace).Get(r.cfg.ServiceCatalogApiServerName, v1.GetOptions{})
		if err != nil {
			return false, err
		}
		ready := deployment.Status.ReadyReplicas
		available := deployment.Status.AvailableReplicas
		if ready >= 1 && available >= 1 {
			return true, nil
		}
		return false, nil
	})
}

func (r *readiness) assertServiceCatalogControllerIsUp() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		deployment, err := r.client.AppsV1beta1().Deployments(r.cfg.ServiceCatalogNamespace).Get(r.cfg.ServiceCatalogControllerServerName, v1.GetOptions{})
		if err != nil {
			return false, err
		}
		ready := deployment.Status.ReadyReplicas
		available := deployment.Status.AvailableReplicas
		if ready >= 1 && available >= 1 {
			return true, nil
		}
		return false, nil
	})
}

func (r *readiness) assertTestBrokerIsReady() error {
	klog.Info("Make sure TestBroker is up")
	if err := r.assertTestBrokerIsUp(); err != nil {
		return errors.Wrap(err, "failed during waiting for TestBroker")
	}
	klog.Info("TestBroker is ready")

	return nil
}

func (r *readiness) assertTestBrokerIsUp() error {
	return wait.Poll(waitInterval, timeoutInterval, func() (done bool, err error) {
		deployment, err := r.client.AppsV1beta1().Deployments(r.cfg.TestBrokerNamespace).Get(r.cfg.TestBrokerName, v1.GetOptions{})
		if err != nil {
			return false, err
		}
		ready := deployment.Status.ReadyReplicas
		available := deployment.Status.AvailableReplicas
		if ready >= 1 && available >= 1 {
			return true, nil
		}
		return false, nil
	})
}
