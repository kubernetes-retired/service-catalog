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
	"io/ioutil"
	"strings"

	sc "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sClientSet "k8s.io/client-go/kubernetes"
	admissionregistrationv1beta1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"
)

// Service provides methods (Backup and Restore) to perform a migration from API Server version (0.2.x) to CRDs version (0.3.0).
type Service struct {
	storagePath      string
	releaseNamespace string
	apiserverName    string

	admInterface  admissionregistrationv1beta1.AdmissionregistrationV1beta1Interface
	appInterface  appsv1.AppsV1Interface
	coreInterface corev1.CoreV1Interface
	scInterface   v1beta1.ServicecatalogV1beta1Interface

	marshaller   func(interface{}) ([]byte, error)
	unmarshaller func([]byte, interface{}) error
}

// NewMigrationService creates a new instance of a Service
func NewMigrationService(scInterface v1beta1.ServicecatalogV1beta1Interface, storagePath string, releaseNamespace string, apiserverName string, k8sclient *k8sClientSet.Clientset) *Service {
	return &Service{
		storagePath:      storagePath,
		releaseNamespace: releaseNamespace,
		apiserverName:    apiserverName,

		admInterface:  k8sclient.AdmissionregistrationV1beta1(),
		appInterface:  k8sclient.AppsV1(),
		coreInterface: k8sclient.CoreV1(),
		scInterface:   scInterface,

		marshaller: yaml.Marshal,
		unmarshaller: func(b []byte, obj interface{}) error {
			return yaml.Unmarshal(b, obj)
		},
	}
}

// ServiceCatalogResources aggregates all Service Catalog resources
type ServiceCatalogResources struct {
	clusterServiceBrokers []sc.ClusterServiceBroker
	serviceBrokers        []sc.ServiceBroker
	serviceInstances      []sc.ServiceInstance
	serviceBindings       []sc.ServiceBinding
	serviceClasses        []sc.ServiceClass
	servicePlans          []sc.ServicePlan
	clusterServiceClasses []sc.ClusterServiceClass
	clusterServicePlans   []sc.ClusterServicePlan
}

const (
	serviceBrokerFilePrefix        = "servicebroker"
	clusterServiceBrokerFilePrefix = "clusterservicebroker"
	serviceInstanceFilePrefix      = "serviceinstance"
	serviceBindingFilePrefix       = "servicebinding"

	serviceClassFilePrefix        = "serviceclass"
	servicePlanFilePrefix         = "serviceplan"
	clusterServiceClassFilePrefix = "clusterserviceclass"
	clusterServicePlanFilePrefix  = "clusterserviceplan"
)

// bindingControllerKind contains the schema.GroupVersionKind for this controller type.
var bindingControllerKind = sc.SchemeGroupVersion.WithKind("ServiceBinding")

func (r *ServiceCatalogResources) writeMetadata(b *strings.Builder, m metav1.ObjectMeta) {
	b.WriteString("\n\t")
	b.WriteString(m.Namespace)
	b.WriteString("/")
	b.WriteString(m.Name)
}

func (m *Service) loadResource(filename string, obj interface{}) error {
	b, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", m.storagePath, filename))
	if err != nil {
		return errors.Wrapf(err, "while reading file %s/%s", m.storagePath, filename)
	}
	err = m.unmarshaller(b, obj)
	if err != nil {
		return errors.Wrapf(err, "while unmarshalling file %s/%s", m.storagePath, filename)
	}
	return nil
}

func (m *Service) adjustOwnerReference(om *metav1.ObjectMeta, uidMap map[string]types.UID) {
	if len(om.OwnerReferences) > 0 {
		om.OwnerReferences[0].UID = uidMap[om.OwnerReferences[0].Name]
	}
}

// IsMigrationRequired checks if current version of Service Catalog needs to be migrated
func (m *Service) IsMigrationRequired() (bool, error) {
	_, err := m.appInterface.Deployments(m.releaseNamespace).Get(m.apiserverName, metav1.GetOptions{})
	switch {
	case err == nil:
	case apiErrors.IsNotFound(err):
		return false, nil
	default:
		return false, fmt.Errorf("other type of error: %s", err)
	}
	return true, nil
}

// Restore restores Service Catalog resources and adds necessary owner reference to all secrets pointed by service bindings.
func (m *Service) Restore(res *ServiceCatalogResources) error {
	klog.Infof("Applying %d service brokers", len(res.serviceBrokers))
	for _, sb := range res.serviceBrokers {
		sb.RecalculatePrinterColumnStatusFields()
		sb.ResourceVersion = ""
		created, err := m.scInterface.ServiceBrokers(sb.Namespace).Create(&sb)
		if err != nil {
			return err
		}

		created.Status = sb.Status
		_, err = m.scInterface.ServiceBrokers(sb.Namespace).UpdateStatus(created)
		if err != nil {
			return err
		}
	}

	csbNameToUIDMap := map[string]types.UID{}
	klog.Infof("Applying %d cluster service brokers", len(res.clusterServiceBrokers))
	for _, sb := range res.clusterServiceBrokers {
		sb.RecalculatePrinterColumnStatusFields()
		sb.ResourceVersion = ""
		created, err := m.scInterface.ClusterServiceBrokers().Create(&sb)
		if err != nil {
			return err
		}

		created.Status = sb.Status
		_, err = m.scInterface.ClusterServiceBrokers().UpdateStatus(created)
		if err != nil {
			return err
		}
		csbNameToUIDMap[sb.Name] = created.UID
	}

	klog.Infof("Applying %d service classes", len(res.serviceClasses))
	for _, sc := range res.serviceClasses {
		sc.ResourceVersion = ""
		sc.UID = ""
		created, err := m.scInterface.ServiceClasses(sc.Namespace).Create(&sc)
		if err != nil {
			return err
		}

		created.Status = sc.Status
		_, err = m.scInterface.ServiceClasses(sc.Namespace).UpdateStatus(created)
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d cluster service classes", len(res.clusterServiceClasses))
	for _, csc := range res.clusterServiceClasses {
		csc.ResourceVersion = ""
		csc.UID = ""
		csc.SelfLink = ""
		m.adjustOwnerReference(&csc.ObjectMeta, csbNameToUIDMap)
		created, err := m.scInterface.ClusterServiceClasses().Create(&csc)
		if err != nil {
			return err
		}

		created.Status = csc.Status
		_, err = m.scInterface.ClusterServiceClasses().UpdateStatus(created)
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d service plans", len(res.servicePlans))
	for _, sp := range res.servicePlans {
		sp.ResourceVersion = ""
		sp.UID = ""
		created, err := m.scInterface.ServicePlans(sp.Namespace).Create(&sp)
		if err != nil {
			return err
		}

		created.Status = sp.Status
		_, err = m.scInterface.ServicePlans(sp.Namespace).UpdateStatus(created)
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d cluster service plans", len(res.clusterServicePlans))
	for _, csp := range res.clusterServicePlans {
		csp.ResourceVersion = ""
		csp.UID = ""
		m.adjustOwnerReference(&csp.ObjectMeta, csbNameToUIDMap)
		created, err := m.scInterface.ClusterServicePlans().Create(&csp)
		if err != nil {
			return err
		}

		created.Status = csp.Status
		_, err = m.scInterface.ClusterServicePlans().UpdateStatus(created)
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d service instances", len(res.serviceInstances))
	for _, si := range res.serviceInstances {
		si.RecalculatePrinterColumnStatusFields()
		si.ResourceVersion = ""

		instance := si.DeepCopy()

		// ServiceInstance must not have class/plan refs when it is created
		// These fields must be filled using an update
		si.Spec.ClusterServiceClassRef = nil
		si.Spec.ClusterServicePlanRef = nil
		si.Spec.ServiceClassRef = nil
		si.Spec.ServicePlanRef = nil
		created, err := m.scInterface.ServiceInstances(si.Namespace).Create(&si)
		if err != nil {
			return err
		}

		created.Spec.ClusterServiceClassRef = instance.Spec.ClusterServiceClassRef
		created.Spec.ClusterServicePlanRef = instance.Spec.ClusterServicePlanRef
		created.Spec.ServiceClassRef = instance.Spec.ServiceClassRef
		created.Spec.ServicePlanRef = instance.Spec.ServicePlanRef

		updated, err := m.scInterface.ServiceInstances(si.Namespace).Update(created)
		if err != nil {
			return err
		}

		updated.Status = si.Status
		updated.Status.ObservedGeneration = updated.Generation
		updated, err = m.scInterface.ServiceInstances(si.Namespace).UpdateStatus(updated)
		if err != nil {
			return err
		}
	}

	klog.Infof("Applying %d service bindings", len(res.serviceInstances))
	for _, sb := range res.serviceBindings {
		sb.RecalculatePrinterColumnStatusFields()
		sb.ResourceVersion = ""
		created, err := m.scInterface.ServiceBindings(sb.Namespace).Create(&sb)
		if err != nil {
			return err
		}

		created.Status = sb.Status
		_, err = m.scInterface.ServiceBindings(sb.Namespace).UpdateStatus(created)
		if err != nil {
			return err
		}

		m.AddOwnerReferenceToSecret(created)
	}

	return nil
}

// LoadResources loads Service Catalog resources from files.
func (m *Service) LoadResources() (*ServiceCatalogResources, error) {
	files, err := ioutil.ReadDir(m.storagePath)
	if err != nil {
		return nil, err
	}

	var serviceBrokers []sc.ServiceBroker
	for _, file := range files {
		if strings.HasPrefix(file.Name(), serviceBrokerFilePrefix) {
			var obj sc.ServiceBroker
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			serviceBrokers = append(serviceBrokers, obj)
		}
	}

	var clusterServiceBrokers []sc.ClusterServiceBroker
	for _, file := range files {
		if strings.HasPrefix(file.Name(), clusterServiceBrokerFilePrefix) {
			var obj sc.ClusterServiceBroker
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			clusterServiceBrokers = append(clusterServiceBrokers, obj)
		}
	}

	var serviceInstances []sc.ServiceInstance
	for _, file := range files {
		if strings.HasPrefix(file.Name(), serviceInstanceFilePrefix) {
			var obj sc.ServiceInstance
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			serviceInstances = append(serviceInstances, obj)
		}
	}

	var serviceBinding []sc.ServiceBinding
	for _, file := range files {
		if strings.HasPrefix(file.Name(), serviceBindingFilePrefix) {
			var obj sc.ServiceBinding
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			serviceBinding = append(serviceBinding, obj)
		}
	}

	var serviceClasses []sc.ServiceClass
	for _, file := range files {
		if strings.HasPrefix(file.Name(), serviceClassFilePrefix) {
			var obj sc.ServiceClass
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			serviceClasses = append(serviceClasses, obj)
		}
	}

	var servicePlans []sc.ServicePlan
	for _, file := range files {
		if strings.HasPrefix(file.Name(), servicePlanFilePrefix) {
			var obj sc.ServicePlan
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			servicePlans = append(servicePlans, obj)
		}
	}

	var clusterServiceClasses []sc.ClusterServiceClass
	for _, file := range files {
		if strings.HasPrefix(file.Name(), clusterServiceClassFilePrefix) {
			var obj sc.ClusterServiceClass
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			clusterServiceClasses = append(clusterServiceClasses, obj)
		}
	}

	var clusterServicePlans []sc.ClusterServicePlan
	for _, file := range files {
		if strings.HasPrefix(file.Name(), clusterServicePlanFilePrefix) {
			var obj sc.ClusterServicePlan
			err := m.loadResource(file.Name(), &obj)
			if err != nil {
				return nil, err
			}
			clusterServicePlans = append(clusterServicePlans, obj)
		}
	}

	return &ServiceCatalogResources{
		serviceBrokers:        serviceBrokers,
		serviceInstances:      serviceInstances,
		serviceBindings:       serviceBinding,
		clusterServiceBrokers: clusterServiceBrokers,
		serviceClasses:        serviceClasses,
		servicePlans:          servicePlans,
		clusterServiceClasses: clusterServiceClasses,
		clusterServicePlans:   clusterServicePlans,
	}, nil
}

// Cleanup deletes all given resources
func (m *Service) Cleanup(resources *ServiceCatalogResources) error {
	klog.Infoln("Cleaning up Service Catalog Resources")
	for _, obj := range resources.serviceBindings {
		err := m.scInterface.ServiceBindings(obj.Namespace).Delete(obj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, obj := range resources.serviceInstances {
		err := m.scInterface.ServiceInstances(obj.Namespace).Delete(obj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, obj := range resources.serviceClasses {
		err := m.scInterface.ServiceClasses(obj.Namespace).Delete(obj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, obj := range resources.clusterServiceClasses {
		err := m.scInterface.ClusterServiceClasses().Delete(obj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, obj := range resources.servicePlans {
		err := m.scInterface.ServicePlans(obj.Namespace).Delete(obj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, obj := range resources.clusterServicePlans {
		err := m.scInterface.ClusterServicePlans().Delete(obj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, obj := range resources.serviceBrokers {
		err := m.scInterface.ServiceBrokers(obj.Namespace).Delete(obj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, obj := range resources.clusterServiceBrokers {
		err := m.scInterface.ClusterServiceBrokers().Delete(obj.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	klog.Infoln("...done")
	return nil
}

func (m *Service) backupResource(obj interface{}, filePrefix string, uid types.UID) error {
	const perm = 0644
	b, err := m.marshaller(obj)
	if err != nil {
		return errors.Wrapf(err, "while marshalling file %s/%s-%s", m.storagePath, filePrefix, uid)
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s-%s", m.storagePath, filePrefix, uid), b, perm)
	if err != nil {
		return errors.Wrapf(err, "while writing file %s/%s-%s", m.storagePath, filePrefix, uid)
	}
	return nil
}

// BackupResources saves all Service Catalog resources to files.
func (m *Service) BackupResources() (*ServiceCatalogResources, error) {
	klog.Infoln("Saving resources")
	serviceBrokers, err := m.scInterface.ServiceBrokers(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, sb := range serviceBrokers.Items {
		err := m.backupResource(&sb, serviceBrokerFilePrefix, sb.UID)
		if err != nil {
			return nil, err
		}
	}

	clusterServiceBrokers, err := m.scInterface.ClusterServiceBrokers().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, csb := range clusterServiceBrokers.Items {
		err := m.backupResource(&csb, clusterServiceBrokerFilePrefix, csb.UID)
		if err != nil {
			return nil, err
		}
	}

	serviceClasses, err := m.scInterface.ServiceClasses(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, sc := range serviceClasses.Items {
		err := m.backupResource(&sc, serviceClassFilePrefix, sc.UID)
		if err != nil {
			return nil, err
		}
	}

	clusterServiceClasses, err := m.scInterface.ClusterServiceClasses().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, csc := range clusterServiceClasses.Items {
		err := m.backupResource(&csc, clusterServiceClassFilePrefix, csc.UID)
		if err != nil {
			return nil, err
		}
	}

	servicePlans, err := m.scInterface.ServicePlans(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, sp := range servicePlans.Items {
		err := m.backupResource(&sp, servicePlanFilePrefix, sp.UID)
		if err != nil {
			return nil, err
		}
	}

	clusterServicePlans, err := m.scInterface.ClusterServicePlans().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, csp := range clusterServicePlans.Items {
		err := m.backupResource(&csp, clusterServicePlanFilePrefix, csp.UID)
		if err != nil {
			return nil, err
		}
	}

	serviceInstances, err := m.scInterface.ServiceInstances(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, si := range serviceInstances.Items {
		err := m.backupResource(&si, serviceInstanceFilePrefix, si.UID)
		if err != nil {
			return nil, err
		}
	}

	serviceBindings, err := m.scInterface.ServiceBindings(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, sb := range serviceBindings.Items {
		err := m.backupResource(&sb, serviceBindingFilePrefix, sb.UID)
		if err != nil {
			return nil, err
		}
	}

	klog.Infoln("...done")
	return &ServiceCatalogResources{
		clusterServiceBrokers: clusterServiceBrokers.Items,
		serviceBrokers:        serviceBrokers.Items,
		clusterServiceClasses: clusterServiceClasses.Items,
		serviceClasses:        serviceClasses.Items,
		clusterServicePlans:   clusterServicePlans.Items,
		servicePlans:          servicePlans.Items,
		serviceInstances:      serviceInstances.Items,
		serviceBindings:       serviceBindings.Items,
	}, nil
}

// AddOwnerReferenceToSecret updates a secret (referenced in the given ServiceBinding) by adding proper owner reference
func (m *Service) AddOwnerReferenceToSecret(sb *sc.ServiceBinding) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		secret, err := m.coreInterface.Secrets(sb.Namespace).Get(sb.Spec.SecretName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		secret.OwnerReferences = []metav1.OwnerReference{
			*metav1.NewControllerRef(sb, bindingControllerKind),
		}
		_, err = m.coreInterface.Secrets(sb.Namespace).Update(secret)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// RemoveOwnerReferenceFromSecrets removes owner references from secrets created for service bindings.
func (m *Service) RemoveOwnerReferenceFromSecrets() error {
	klog.Info("Removing owner referneces from secrets")
	serviceBindings, err := m.scInterface.ServiceBindings(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sb := range serviceBindings.Items {
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			secret, err := m.coreInterface.Secrets(sb.Namespace).Get(sb.Spec.SecretName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			secret.OwnerReferences = []metav1.OwnerReference{}
			_, err = m.coreInterface.Secrets(sb.Namespace).Update(secret)
			return err
		})
		if err != nil {
			return err
		}
	}
	klog.Infoln("...done")
	return nil
}
