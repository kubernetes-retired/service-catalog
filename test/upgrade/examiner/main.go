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

package main

import (
	"flag"
	"fmt"

	"github.com/kubernetes-sigs/service-catalog/test/upgrade/examiner/internal/clientutil"
	"github.com/kubernetes-sigs/service-catalog/test/upgrade/examiner/internal/readiness"
	"github.com/kubernetes-sigs/service-catalog/test/upgrade/examiner/internal/runner"
	"github.com/kubernetes-sigs/service-catalog/test/upgrade/examiner/internal/tests/broker"
	"github.com/kubernetes-sigs/service-catalog/test/upgrade/examiner/internal/tests/clusterbroker"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const (
	prepareDataActionName  = "prepareData"
	executeTestsActionName = "executeTests"
)

func registeredTests(cs *clientutil.ClientStorage) map[string]runner.UpgradeTest {
	return map[string]runner.UpgradeTest{
		"test-broker":         broker.NewTestBroker(cs),
		"test-cluster-broker": clusterbroker.NewTestBroker(cs),
	}
}

// Config collects all parameters from env variables
type Config struct {
	Local          bool         `envconfig:"default=true"`
	KubeconfigPath string       `envconfig:"optional"`
	KubeConfig     *rest.Config `envconfig:"-"`
	readiness.ServiceCatalogConfig
}

// ConfigFlag collects all parameters from flags
type ConfigFlag struct {
	Action string
}

func main() {
	// setup all configurations: envs, flags, stop channel
	flg := readFlags()
	cfg, err := readConfig()
	fatalOnError(err, "while create config")
	stop := server.SetupSignalHandler()

	// create client storage - struct with all required clients
	cs, err := clientutil.NewClientStorage(cfg.KubeConfig)
	fatalOnError(err, "while create kubernetes client storage")

	// get tests
	upgradeTests := registeredTests(cs)

	// get runner
	testRunner, err := runner.NewTestRunner(cs.KubernetesClient().CoreV1().Namespaces(), upgradeTests)
	fatalOnError(err, "while creating test runner")

	// launch runner
	switch flg.Action {
	case prepareDataActionName:
		// make sure ServiceCatalog and TestBroker are ready
		ready := readiness.NewReadiness(cs, cfg.ServiceCatalogConfig)
		err = ready.TestEnvironmentIsReady()
		fatalOnError(err, "while check ServiceCatalog/TestBroker readiness")

		// prepare data for tests
		err := testRunner.PrepareData(stop)
		fatalOnError(err, "while executing prepare data for all registered tests")
	case executeTestsActionName:
		err := testRunner.ExecuteTests(stop)
		fatalOnError(err, "while executing tests for all registered tests")
	default:
		klog.Fatalf("Unrecognized runner action. Allowed actions: %s or %s.", prepareDataActionName, executeTestsActionName)
	}
}

func fatalOnError(err error, context string) {
	if err != nil {
		klog.Fatalf("%s: %v", context, err)
	}
}

func readFlags() ConfigFlag {
	var action string

	flag.StringVar(&action, "action", "", fmt.Sprintf("Define what kind of action runner should execute. Possible values: %s or %s", prepareDataActionName, executeTestsActionName))

	err := flag.Set("logtostderr", "true")
	fatalOnError(err, "while set flag logtostderr")

	err = flag.Set("alsologtostderr", "true")
	fatalOnError(err, "while set flag alsologtostderr")

	flag.Parse()

	return ConfigFlag{
		Action: action,
	}
}

func readConfig() (Config, error) {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err, "while reading configuration from environment variables")

	if cfg.Local && cfg.KubeconfigPath == "" {
		return cfg, errors.New("KubeconfigPath is required for local mode")
	}

	if cfg.Local {
		cfg.KubeConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeconfigPath)
	} else {
		cfg.KubeConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		return cfg, errors.Wrap(err, "while get kubernetes client config")
	}

	return cfg, nil
}
