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
	"fmt"
	"os"

	"github.com/kubernetes-sigs/service-catalog/pkg/cleaner"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-sigs/service-catalog/pkg/migration"
	k8sClientSet "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const (
	blockerBaseName string = "service-catalog-migration-blocker"
)

// RunCommand executes migration action
func RunCommand(opt *Options) error {
	if err := opt.Validate(); nil != err {
		return err
	}

	restConfig, err := newRestClientConfig()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client config: %s", err)
	}
	scClient, err := sc.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Service Catalog client: %s", err)
	}
	k8sCli, err := k8sClientSet.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %s", err)
	}
	scInterface := scClient.ServicecatalogV1beta1()

	svc := migration.NewMigrationService(scInterface, opt.StoragePath, opt.ReleaseNamespace, opt.ApiserverName, k8sCli)
	scalingSvc := migration.NewScalingService(opt.ReleaseNamespace, opt.ControllerManagerName, k8sCli.AppsV1())

	switch opt.Action {
	case backupActionName:
		isMigrationRequired, err := svc.IsMigrationRequired()
		if err != nil {
			return err
		}
		if !isMigrationRequired {
			klog.Infoln("Missing Apiserver deployment - skipping the migration")
			return nil
		}

		klog.Infoln("Executing backup action")

		svc.DisableBlocker(blockerBaseName)
		err = svc.EnableBlocker(blockerBaseName)
		if err != nil {
			return err
		}

		// This defer is a fail-safe to clean up in case of any issue in backup process
		// DisableBlocker can be safely called multiple times without generating errors
		defer svc.DisableBlocker(blockerBaseName)

		err = scalingSvc.ScaleDown()
		if err != nil {
			return err
		}

		res, err := svc.BackupResources()
		if err != nil {
			return err
		}

		err = svc.RemoveOwnerReferenceFromSecrets()
		if err != nil {
			return err
		}

		// Blocker has to be disabled cause we are about to remove protected objects
		svc.DisableBlocker(blockerBaseName)

		err = svc.Cleanup(res)
		if err != nil {
			return err
		}

		klog.Infoln("Removing finalizers")
		finalizerCleaner := cleaner.NewFinalizerCleaner(scClient)
		if err = finalizerCleaner.RemoveFinalizers(); err != nil {
			return err
		}

		klog.Infoln("...done")
		return nil
	case restoreActionName:
		klog.Infoln("Executing restore action")
		err := scalingSvc.ScaleDown()
		if err != nil {
			return err
		}

		res, err := svc.LoadResources()
		if err != nil {
			return err
		}

		err = svc.Restore(res)
		if err != nil {
			return err
		}

		err = scalingSvc.ScaleUp()
		if err != nil {
			return err
		}
	case deployBlockerActionName:
		return svc.EnableBlocker(blockerBaseName)
	case undeployBlockerActionName:
		svc.DisableBlocker(blockerBaseName)
	default:
		return fmt.Errorf("unknown action %s", opt.Action)
	}
	return nil
}

func newRestClientConfig() (*restclient.Config, error) {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	klog.V(4).Info("KUBECONFIG not defined, creating in-cluster config")
	return restclient.InClusterConfig()
}
