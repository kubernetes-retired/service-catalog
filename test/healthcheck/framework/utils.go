/*
Copyright 2018 The Kubernetes Authors.

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

package framework

import (
	"fmt"
	"runtime"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	// How often to poll for conditions
	Poll = 2 * time.Second

	// Default time to wait for operations to complete
	defaultTimeout = 30 * time.Second
)

// RestclientConfig builds a Config object
func RestclientConfig(config, context string) (*api.Config, error) {
	if config == "" {
		return nil, fmt.Errorf("Config file must be specified to load client config")
	}
	c, err := clientcmd.LoadFromFile(config)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err.Error())
	}
	if context != "" {
		c.CurrentContext = context
	}
	return c, nil
}

// LoadConfig parses the config and context and returns a new Config
func LoadConfig(config, context string) (*rest.Config, error) {
	c, err := RestclientConfig(config, context)
	if err != nil {
		return nil, err
	}
	return clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{}).ClientConfig()
}

// CreateKubeNamespace create a new K8s namespace with a unique name
func CreateKubeNamespace(c kubernetes.Interface) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("svc-catalog-health-check-%v-", uuid.NewUUID()),
		},
	}

	// Be robust about making the namespace creation call.
	var got *corev1.Namespace
	err := wait.PollImmediate(Poll, defaultTimeout, func() (bool, error) {
		var err error
		got, err = c.CoreV1().Namespaces().Create(ns)
		if err != nil {
			glog.Errorf("Unexpected error while creating namespace: %v", err)
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return nil, logErrorf("Error creating test namespace: %v", err.Error())
	}
	return got, nil
}

// DeleteKubeNamespace deletes the specified K8s namespace
func DeleteKubeNamespace(c kubernetes.Interface, namespace string) error {
	return c.CoreV1().Namespaces().Delete(namespace, nil)
}

// WaitForEndpoint waits for 'defaultTimeout' interval for an enpoint to be available
func WaitForEndpoint(c kubernetes.Interface, namespace, name string) error {
	return wait.PollImmediate(Poll, defaultTimeout, endpointAvailable(c, namespace, name))
}

func endpointAvailable(c kubernetes.Interface, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		endpoint, err := c.CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if apierrs.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		if len(endpoint.Subsets) == 0 || len(endpoint.Subsets[0].Addresses) == 0 {
			return false, nil
		}

		return true, nil
	}
}

// logErrorf creates a new error using msg and param for the formated message.
// The message is logged and the new error returned.  This function attempts to
// log the location of the caller (file name & line number) so as to maintain
// context of where the error occured
func logErrorf(msg, param string) error {
	_, file, line, _ := runtime.Caller(1)

	// only use the last 30 characters
	context := len(file) - 30
	if context < 0 {
		context = 0
	}
	partialFileName := file[context:]
	format := fmt.Sprintf("...%s:%d: %v", partialFileName, line, msg)
	e := fmt.Errorf(format, param)
	glog.Error(e)
	return e
}
