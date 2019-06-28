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

package runner

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	createdByLabelName = "svc-cat.io/created-by"
	// negative value for regex used for name validation in k8s:
	// https://github.com/kubernetes/apimachinery/blob/98853ca904e81a25e2000cae7f077dc30f81b85f/pkg/util/validation/validation.go#L110
	regexSanitize = "[^a-z0-9]([^-a-z0-9]*[^a-z0-9])?"
)

// taskFn is a signature for task function.
// Required to unify the way how UpgradeTest methods are executed.
type taskFn func(stopCh <-chan struct{}, namespace string) error

// NamespaceCreator has methods required to create k8s ns.
type NamespaceCreator interface {
	Create(*v1.Namespace) (*v1.Namespace, error)
}

// TestRunner executes registered tests
type TestRunner struct {
	nsCreator     NamespaceCreator
	tests         map[string]UpgradeTest
	sanitizeRegex *regexp.Regexp
}

// NewTestRunner is a constructor for TestRunner
func NewTestRunner(nsCreator NamespaceCreator, tests map[string]UpgradeTest) (*TestRunner, error) {
	sanitizeRegex, err := regexp.Compile(regexSanitize)
	if err != nil {
		return nil, errors.Wrap(err, "while compiling sanitize regexp")
	}

	return &TestRunner{
		nsCreator:     nsCreator,
		tests:         tests,
		sanitizeRegex: sanitizeRegex,
	}, nil
}

// PrepareData prepares data for all registered upgrade tests
func (r *TestRunner) PrepareData(stopCh <-chan struct{}) error {
	var count int

	for name, test := range r.tests {
		failed := r.executeTaskFunc(test.CreateResources, stopCh, "CreateResources", name, true)
		if failed {
			count++
		}
	}

	if count > 0 {
		return fmt.Errorf("executed %d task and %d of them failed", len(r.tests), count)
	}

	return nil
}

// ExecuteTests executes all registered tests
func (r *TestRunner) ExecuteTests(stopCh <-chan struct{}) error {
	var count int

	for name, test := range r.tests {
		failed := r.executeTaskFunc(test.TestResources, stopCh, "TestResources", name, false)
		if failed {
			count++
		}
	}

	if count > 0 {
		return fmt.Errorf("executed %d task and %d of them failed", len(r.tests), count)
	}

	return nil
}

func (r *TestRunner) executeTaskFunc(taskHandler taskFn, stopCh <-chan struct{}, header, taskName string, createNs bool) bool {
	fullHeader := fmt.Sprintf("[%s: %s]", header, taskName)

	if r.shutdownRequested(stopCh) {
		klog.Infof("Stop channel called. Not executing %s", fullHeader)
		return true
	}

	klog.Infof("%s Starting execution", fullHeader)

	nsName := r.sanitizedNamespaceName(taskName)
	if createNs {
		if err := r.ensureNamespaceExists(nsName); err != nil {
			klog.Errorf("Cannot create namespace %q: %v", nsName, err)
			return true
		}
	}

	startTime := time.Now()

	if err := taskHandler(stopCh, nsName); err != nil {
		klog.Errorf("%s End with error [start time: %v, duration: %v]: %v", fullHeader, startTime.Format(time.RFC1123), time.Since(startTime), err)
		return true
	}

	klog.Infof("%s End with success [start time: %v, duration: %v]", fullHeader, startTime.Format(time.RFC1123), time.Since(startTime))

	return false
}

// sanitizedNamespaceName returns sanitized name base on rules from this site:
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
func (r *TestRunner) sanitizedNamespaceName(nameToSanitize string) string {
	nsName := strings.ToLower(nameToSanitize)
	nsName = r.sanitizeRegex.ReplaceAllString(nsName, "-")

	if len(nsName) > 253 {
		nsName = nsName[:253]
	}

	return nsName
}

func (r *TestRunner) ensureNamespaceExists(name string) error {
	_, err := r.nsCreator.Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"env":              "true",
				createdByLabelName: "upgrade-test",
			},
		},
	})

	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *TestRunner) shutdownRequested(stopCh <-chan struct{}) bool {
	select {
	case <-stopCh:
		return true
	default:
	}
	return false
}
