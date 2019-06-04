/*
Copyright 2015 The Kubernetes Authors.

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

package e2e

import (
	"testing"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/kubernetes-sigs/service-catalog/test/e2e/framework"
	"k8s.io/klog"
)

func init() {
	framework.RegisterParseFlags()

	if "" == framework.TestContext.KubeConfig {
		klog.Fatalf("environment variable %v must be set", clientcmd.RecommendedConfigPathEnvVar)
	}
	if "" == framework.TestContext.ServiceCatalogConfig {
		klog.Fatalf("environment variable %v must be set", framework.RecommendedConfigPathEnvVar)
	}
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}
