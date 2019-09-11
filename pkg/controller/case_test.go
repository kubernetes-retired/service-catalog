// +build integration

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

package controller_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"sync"

	osb "github.com/kubernetes-sigs/go-open-service-broker-client/v2"
	fakeosb "github.com/kubernetes-sigs/go-open-service-broker-client/v2/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	fakesc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	scinterface "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	scinformers "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kubernetes-sigs/service-catalog/pkg/controller"
	scfeatures "github.com/kubernetes-sigs/service-catalog/pkg/features"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhook/servicecatalog/clusterserviceclass/mutation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	k8sinformers "k8s.io/client-go/informers"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

const (
	testNamespace                         = "test-ns"
	testClusterServiceBrokerName          = "test-clusterservicebroker"
	testClusterServiceClassName           = "test-clusterserviceclass"
	testClusterServicePlanName            = "test-clusterserviceplan"
	testOtherClusterServicePlanName       = "test-otherclusterserviceplan"
	testServiceInstanceName               = "service-instance"
	testClassExternalID                   = "clusterserviceclass-12345"
	testPlanExternalID                    = "34567"
	testOtherPlanExternalID               = "76543"
	testNonbindablePlanExternalID         = "nb34567"
	testNonbindableClusterServicePlanName = "test-nonbindable-plan"
	testExternalID                        = "9737b6ed-ca95-4439-8219-c53fcad118ab"
	testBindingName                       = "test-binding"
	testServiceBindingGUID                = "bguid"
	authSecretName                        = "basic-secret-name"
	testUsername                          = "some-user"
	secretNameWithParameters              = "secret-name"
	secretKeyWithParameters               = "secret-key"
	otherSecretNameWithParameters         = "other-secret-name"
	otherSecretKeyWithParameters          = "other-secret-key"
	testDashboardURL                      = "http://test-dashboard.example.com"

	pollingInterval = 50 * time.Millisecond
	pollingTimeout  = 8 * time.Second
)

// controllerTest provides helper methods to create and verify ServiceCatalog resources.
// Every test case needs a new instance of the controllerTest.
type controllerTest struct {
	// resource clientsets and interfaces
	scInterface      scinterface.ServicecatalogV1beta1Interface
	k8sClient        *fakek8s.Clientset
	fakeOSBClient    *fakeosb.FakeClient
	catalogReactions []fakeosb.CatalogReaction
	osbClientCfg     *osb.ClientConfiguration
	stopCh           chan struct{}

	serviceBindingHandler        *serviceBindingHandler
	serviceBindingInformerStopCh chan struct{}
}

// newControllerTest creates a controllerTest instance with a ready to test running Controller
func newControllerTest(t *testing.T) *controllerTest {
	k8sClient := fakek8s.NewSimpleClientset()
	k8sClient.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
			UID:  uuid.NewUUID(), // fake clientset doesn't  duplicate server-side behavior like uid assignment
		},
	})

	fakeOSBClient := fakeosb.NewFakeClient(fixtureHappyPathBrokerClientConfig())

	coreInformerFactory := k8sinformers.NewSharedInformerFactory(k8sClient, time.Minute)
	coreInformers := coreInformerFactory.Core()

	scClient := fakesc.NewSimpleClientset()
	informerFactory := scinformers.NewSharedInformerFactory(scClient, 0)
	serviceCatalogSharedInformers := informerFactory.Servicecatalog().V1beta1()

	clusterServiceClassInformer := serviceCatalogSharedInformers.ClusterServiceClasses()
	plansInformer := serviceCatalogSharedInformers.ClusterServicePlans()

	testCase := &controllerTest{
		scInterface:      scClient.ServicecatalogV1beta1(),
		k8sClient:        k8sClient,
		fakeOSBClient:    fakeOSBClient,
		catalogReactions: []fakeosb.CatalogReaction{},
	}

	// wrap the ClientFunc with a helper which saves last used OSG Client Config (it can be asserted in the test)
	brokerClFunc := testCase.spyOSBClientFunc(fakeosb.ReturnFakeClientFunc(fakeOSBClient))

	fakeRecorder := record.NewFakeRecorder(1)
	// start goroutine which flushes events (prevent hanging recording function)
	go func() {
		for range fakeRecorder.Events {
		}
	}()

	testController, err := controller.NewController(
		k8sClient,
		coreInformers.V1().Secrets(),
		scClient.ServicecatalogV1beta1(),
		serviceCatalogSharedInformers.ClusterServiceBrokers(),
		serviceCatalogSharedInformers.ServiceBrokers(),
		clusterServiceClassInformer,
		serviceCatalogSharedInformers.ServiceClasses(),
		serviceCatalogSharedInformers.ServiceInstances(),
		serviceCatalogSharedInformers.ServiceBindings(),
		plansInformer,
		serviceCatalogSharedInformers.ServicePlans(),
		brokerClFunc,
		24*time.Hour,
		osb.LatestAPIVersion().HeaderValue(),
		fakeRecorder,
		7*24*time.Hour,
		7*24*time.Hour,
		"DefaultClusterIDConfigMapName",
		"DefaultClusterIDConfigMapNamespace",
		60*time.Second,
	)
	if err != nil {
		t.Fatal(err)
	}

	sbHandler := &serviceBindingHandler{}
	serviceBindingInformer := serviceCatalogSharedInformers.ServiceBindings().Informer()
	serviceBindingInformer.AddEventHandler(sbHandler)

	// start and sync informers
	testCase.stopCh = make(chan struct{})
	informerFactory.Start(testCase.stopCh)
	coreInformerFactory.Start(testCase.stopCh)
	informerFactory.WaitForCacheSync(testCase.stopCh)
	coreInformerFactory.WaitForCacheSync(testCase.stopCh)

	// start the controller
	go testController.Run(1, testCase.stopCh)

	testCase.serviceBindingHandler = sbHandler
	return testCase
}

// serviceBindingHandler handles notifications for events that happen to ServiceBinding.
type serviceBindingHandler struct {
	sync.Mutex

	onUpdate func(old, new *v1beta1.ServiceBinding)
}

// OnAdd handles ServiceBinding add action.
// Empty implementation just to fulfill cache.ResourceEventHandler interface
func (*serviceBindingHandler) OnAdd(obj interface{}) {
}

// OnDelete handles ServiceBinding delete action.
// Empty implementation just to fulfill cache.ResourceEventHandler interface
func (*serviceBindingHandler) OnDelete(obj interface{}) {
}

// OnUpdate handles ServiceBinding update action and execute injected onUpdate function.
func (h *serviceBindingHandler) OnUpdate(old, obj interface{}) {
	h.Lock()
	defer h.Unlock()
	if h.onUpdate == nil {
		return
	}
	newSb, ok := obj.(*v1beta1.ServiceBinding)
	if !ok {
		return
	}
	oldSb, ok := old.(*v1beta1.ServiceBinding)
	if !ok {
		return
	}
	h.onUpdate(oldSb, newSb)
}

// SetServiceBindingOnChangeListener sets callback function which is called
// when ServiceBinding was changed.
func (ct *controllerTest) SetServiceBindingOnChangeListener(onUpdate func(old, new *v1beta1.ServiceBinding)) {
	ct.serviceBindingHandler.Lock()
	defer ct.serviceBindingHandler.Unlock()
	ct.serviceBindingHandler.onUpdate = onUpdate
}

// TearDown performs cleanup for controllerTest instance.
func (ct *controllerTest) TearDown() {
	close(ct.stopCh)
}

// EnableAsyncInstanceProvisioning configures all fake OSB client provision
// responses with async flag
func (ct *controllerTest) EnableAsyncInstanceProvisioning() {
	ct.fakeOSBClient.ProvisionReaction.(*fakeosb.ProvisionReaction).Response.Async = true
}

// EnableAsyncInstanceUpdate configures all fake OSB client update
// responses with async flag
func (ct *controllerTest) EnableAsyncInstanceUpdate() {
	ct.fakeOSBClient.UpdateInstanceReaction.(*fakeosb.UpdateInstanceReaction).Response.Async = true
}

// AsyncForInstanceUpdate configures all fake OSB client update
// responses with async flag
func (ct *controllerTest) AsyncForInstanceUpdate() {
	ct.fakeOSBClient.UpdateInstanceReaction.(*fakeosb.UpdateInstanceReaction).Response.Async = true
}

// EnableAsyncInstanceDeprovisioning configures all fake OSB client deprovision
// responses with async flag
func (ct *controllerTest) EnableAsyncInstanceDeprovisioning() {
	ct.fakeOSBClient.DeprovisionReaction.(*fakeosb.DeprovisionReaction).Response.Async = true
}

// EnableAsyncUnbind configures fake OSB client unbind operation responses with async flag
func (ct *controllerTest) EnableAsyncUnbind() {
	ct.fakeOSBClient.UnbindReaction.(*fakeosb.UnbindReaction).Response.Async = true
}

// EnableAsyncBind configures fake OSB client bind operation responses with async flag
func (ct *controllerTest) EnableAsyncBind() {
	ct.fakeOSBClient.BindReaction.(*fakeosb.BindReaction).Response.Async = true
}

// DisableAsyncUnbind configures fake OSB client unbind operation responses with async flag false
func (ct *controllerTest) DisableAsyncUnbind() {
	ct.fakeOSBClient.UnbindReaction.(*fakeosb.UnbindReaction).Response.Async = false
}

// DisableAsyncBind configures fake OSB client bind operation responses with async flag false
func (ct *controllerTest) DisableAsyncBind() {
	ct.fakeOSBClient.BindReaction.(*fakeosb.BindReaction).Response.Async = false
}

// AssertOSBBasicAuth verifies the last call to broker whether the correct basic auth credentials was used
func (ct *controllerTest) AssertOSBBasicAuth(t *testing.T, username, password string) {
	require.NotNil(t, ct.osbClientCfg, "OSB Client was not created, wait for broker is ready")
	assert.Equal(t, ct.osbClientCfg.AuthConfig.BasicAuthConfig, &osb.BasicAuthConfig{
		Username: username,
		Password: password,
	})
}

// NumberOfOSBUnbindingCalls returns the total number of OSB unbinding calls
func (ct *controllerTest) NumberOfOSBUnbindingCalls() int {
	return ct.numberOfOSBActionByType(fakeosb.Unbind)
}

// NumberOfOSBBindingCalls returns the total number of OSB binding calls
func (ct *controllerTest) NumberOfOSBBindingCalls() int {
	return ct.numberOfOSBActionByType(fakeosb.Bind)
}

// NumberOfOSBProvisionCalls return the total number of OSB provision calls
func (ct *controllerTest) NumberOfOSBProvisionCalls() int {
	return ct.numberOfOSBActionByType(fakeosb.ProvisionInstance)
}

// NumberOfOSBUpdateCalls return the total number of OSB update calls
func (ct *controllerTest) NumberOfOSBUpdateCalls() int {
	return ct.numberOfOSBActionByType(fakeosb.UpdateInstance)
}

// NumberOfOSBDeprovisionCalls returns the total number of OSB deprovision calls
func (ct *controllerTest) NumberOfOSBDeprovisionCalls() int {
	return ct.numberOfOSBActionByType(fakeosb.DeprovisionInstance)
}

// numberOfOSBActionByType returns total number of given OSB action type
func (ct *controllerTest) numberOfOSBActionByType(actionType fakeosb.ActionType) int {
	actions := ct.fakeOSBClient.Actions()
	counter := 0
	for _, action := range actions {
		if action.Type == actionType {
			counter = counter + 1
		}
	}
	return counter
}

// SetFirstOSBPollLastOperationReactionsInProgress makes the broker
// responses inProgress in first numberOfInProgressResponses calls
func (ct *controllerTest) SetFirstOSBPollLastOperationReactionsInProgress(numberOfInProgressResponses int) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	numberOfPolls := 0
	ct.fakeOSBClient.PollLastOperationReaction = fakeosb.DynamicPollLastOperationReaction(
		func(_ *osb.LastOperationRequest) (*osb.LastOperationResponse, error) {
			numberOfPolls++
			state := osb.StateInProgress
			if numberOfPolls > numberOfInProgressResponses {
				state = osb.StateSucceeded
			}
			return &osb.LastOperationResponse{State: state}, nil
		})
}

// SetOSBPollLastOperationReactionsState makes the broker
// responses with given state
func (ct *controllerTest) SetOSBPollLastOperationReactionsState(state osb.LastOperationState) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	ct.fakeOSBClient.PollLastOperationReaction = &fakeosb.PollLastOperationReaction{
		Response: &osb.LastOperationResponse{State: state},
	}
}

// SetOSBPollBindingLastOperationReactionsState makes the broker
// responses with given state
func (ct *controllerTest) SetOSBPollBindingLastOperationReactionsState(state osb.LastOperationState) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	ct.fakeOSBClient.PollBindingLastOperationReaction = &fakeosb.PollBindingLastOperationReaction{
		Response: &osb.LastOperationResponse{State: state},
	}
}

// SetFirstOSBPollLastOperationReactionsInProgress makes the broker
// responses failed in first numberOfInProgressResponses calls
func (ct *controllerTest) SetFirstOSBPollLastOperationReactionsFailed(numberOfFailedResponses int) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	numberOfPolls := 0
	ct.fakeOSBClient.PollLastOperationReaction = fakeosb.DynamicPollLastOperationReaction(
		func(_ *osb.LastOperationRequest) (*osb.LastOperationResponse, error) {
			numberOfPolls++
			state := osb.StateFailed
			if numberOfPolls > numberOfFailedResponses {
				state = osb.StateSucceeded
			}
			return &osb.LastOperationResponse{State: state}, nil
		})
}

// SetFirstOSBPollBindingLastOperationReactionsFailed makes the broker
// responses failed in first numberOfInProgressResponses calls
func (ct *controllerTest) SetFirstOSBPollBindingLastOperationReactionsFailed(numberOfFailedResponses int) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	numberOfPolls := 0
	ct.fakeOSBClient.PollBindingLastOperationReaction = fakeosb.DynamicPollBindingLastOperationReaction(
		func(_ *osb.BindingLastOperationRequest) (*osb.LastOperationResponse, error) {
			numberOfPolls++
			state := osb.StateFailed
			if numberOfPolls > numberOfFailedResponses {
				state = osb.StateSucceeded
			}
			fmt.Println(state)
			return &osb.LastOperationResponse{State: state}, nil
		})
}

// SetFirstOSBProvisionReactionsHTTPError makes the broker
// responses with error in first numberOfInProgressResponses calls
func (ct *controllerTest) SetFirstOSBProvisionReactionsHTTPError(numberOfErrorResponses int, code int) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	numberOfPolls := 0
	ct.fakeOSBClient.ProvisionReaction = fakeosb.DynamicProvisionReaction(
		func(_ *osb.ProvisionRequest) (*osb.ProvisionResponse, error) {
			numberOfPolls++
			if numberOfPolls > numberOfErrorResponses {
				return &osb.ProvisionResponse{}, nil
			}
			return nil, osb.HTTPStatusCodeError{
				StatusCode: code,
			}
		})
}

// SetFirstOSBUnbindReactionsHTTPError makes the broker
// responses with error in first numberOfErrorResponses calls
func (ct *controllerTest) SetFirstOSBUnbindReactionsHTTPError(numberOfErrorResponses int, code int) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	numberOfPolls := 0
	ct.fakeOSBClient.UnbindReaction = fakeosb.DynamicUnbindReaction(
		func(_ *osb.UnbindRequest) (*osb.UnbindResponse, error) {
			numberOfPolls++
			if numberOfPolls > numberOfErrorResponses {
				return &osb.UnbindResponse{}, nil
			}
			return nil, osb.HTTPStatusCodeError{
				StatusCode: code,
			}
		})
}

// SetOSBBindReactionWithHTTPError configures the broker Bind call response as HTTPStatusCodeError
func (ct *controllerTest) SetOSBBindReactionWithHTTPError(code int) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	ct.fakeOSBClient.BindReaction = &fakeosb.BindReaction{
		Error: osb.HTTPStatusCodeError{
			StatusCode: code,
		},
	}
}

// spyOSBClientFunc wraps the ClientFunc with a helper which saves last used OSG Client Config
func (ct *controllerTest) spyOSBClientFunc(target osb.CreateFunc) osb.CreateFunc {
	return func(osbCfg *osb.ClientConfiguration) (osb.Client, error) {
		ct.osbClientCfg = osbCfg
		return target(osbCfg)
	}
}

// fixClusterServiceBroker returns ClusterServiceBroker with filled in all required field
func (ct *controllerTest) fixClusterServiceBroker() *v1beta1.ClusterServiceBroker {
	return &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: testClusterServiceBrokerName,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL:            "https://broker.example.com",
				RelistBehavior: v1beta1.ServiceBrokerRelistBehaviorDuration,
				RelistDuration: &metav1.Duration{Duration: 15 * time.Minute},
			},
		},
	}
}

// CreateSimpleClusterServiceBroker creates a ClusterServiceBroker used in testing scenarios.
func (ct *controllerTest) CreateSimpleClusterServiceBroker() error {
	_, err := ct.scInterface.ClusterServiceBrokers().Create(ct.fixClusterServiceBroker())
	return err
}

// CreateClusterServiceBrokerWithBasicAuth creates a ClusterServiceBroker with basic auth.
func (ct *controllerTest) CreateClusterServiceBrokerWithBasicAuth() error {
	csb := ct.fixClusterServiceBroker()
	csb.Spec.AuthInfo = &v1beta1.ClusterServiceBrokerAuthInfo{
		Basic: &v1beta1.ClusterBasicAuthConfig{
			SecretRef: &v1beta1.ObjectReference{
				Name:      authSecretName,
				Namespace: testNamespace,
			},
		},
	}
	_, err := ct.scInterface.ClusterServiceBrokers().Create(csb)
	return err
}

// AddServiceClassRestrictionsToBroker updates a broker with a restrictions, which must filter out all existing classes.
func (ct *controllerTest) AddServiceClassRestrictionsToBroker() error {
	classes, err := ct.scInterface.ClusterServiceClasses().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	var restrictions []string
	for _, cl := range classes.Items {
		restrictions = append(restrictions, fmt.Sprintf("name!=%s", cl.Name))
	}

	csb, err := ct.scInterface.ClusterServiceBrokers().Get(testClusterServiceBrokerName, metav1.GetOptions{})
	csb.Spec.CatalogRestrictions = &v1beta1.CatalogRestrictions{
		ServiceClass: restrictions,
	}
	csb.Generation = csb.Generation + 1
	_, err = ct.scInterface.ClusterServiceBrokers().Update(csb)
	return err
}

// CreateServiceInstance creates a ServiceInstance which is used in testing scenarios.
func (ct *controllerTest) CreateServiceInstance() error {
	_, err := ct.scInterface.ServiceInstances(testNamespace).Create(&v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: testServiceInstanceName,
			// added by a Webhook, which is not tested here
			Finalizers: []string{v1beta1.FinalizerServiceCatalog},
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: testClassExternalID,
				ClusterServicePlanExternalName:  testPlanExternalID,
			},
			ExternalID: testExternalID,
			// Plan and Class refs are added by a Webhook, which is not tested here
			ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
				Name: testPlanExternalID,
			},
			ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
				Name: testClassExternalID,
			},
			UserInfo: fixtureUserInfo(),
		},
	})
	return err
}

// UpdateServiceInstanceParameters simulates update process of ServiceInstance parameters
func (ct *controllerTest) UpdateServiceInstanceParameters() error {
	si, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		return nil
	}
	parameters := map[string]interface{}{
		"param-key": "new-param-value",
	}
	marshalledParams, err := json.Marshal(parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters %v : %v", parameters, err)
	}
	si.Spec.Parameters = &runtime.RawExtension{Raw: marshalledParams}
	si.Generation = si.Generation + 1

	_, err = ct.scInterface.ServiceInstances(testNamespace).Update(si)
	return err
}

// Deprovision sets deletion timestamp which is done by K8s in a cluster while ServiceInstance deletion.
func (ct *controllerTest) Deprovision() error {
	si, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	si.DeletionTimestamp = ct.v1Now()
	_, err = ct.scInterface.ServiceInstances(testNamespace).Update(si)
	return err
}

// CreateBinding creates a ServiceBinding which is used in testing scenarios.
func (ct *controllerTest) CreateBinding() error {
	_, err := ct.scInterface.ServiceBindings(testNamespace).Create(&v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  testNamespace,
			Name:       testBindingName,
			Generation: 1,
			Finalizers: []string{v1beta1.FinalizerServiceCatalog}, // set by the Webhook
		},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: v1beta1.LocalObjectReference{
				Name: testServiceInstanceName,
			},
			ExternalID: testServiceBindingGUID,
			SecretName: testBindingName, // set by the webhook
			UserInfo:   fixtureUserInfo(),
		},
	})
	return err
}

// Unbind sets deletion timestamp which is done by K8s in a cluster. It triggers unbinding process.
func (ct *controllerTest) Unbind() error {
	sb, err := ct.scInterface.ServiceBindings(testNamespace).Get(testBindingName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	sb.DeletionTimestamp = ct.v1Now()
	_, err = ct.scInterface.ServiceBindings(testNamespace).Update(sb)
	return err
}

// DeleteBinding removes the ServiceBinding resource.
func (ct *controllerTest) DeleteBinding() error {
	return ct.scInterface.ServiceBindings(testNamespace).Delete(testBindingName, &metav1.DeleteOptions{})
}

// CreateSecretWithBasicAuth creates a secret with credentials
// referenced by a ClusterServiceBroker created by CreateClusterServiceBrokerWithBasicAuth method.
func (ct *controllerTest) CreateSecretWithBasicAuth(username, password string) error {
	_, err := ct.k8sClient.CoreV1().Secrets(testNamespace).Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      authSecretName,
		},
		Data: map[string][]byte{
			v1beta1.BasicAuthUsernameKey: []byte(username),
			v1beta1.BasicAuthPasswordKey: []byte(password),
		},
	})
	return err
}

// UpdateSecretWithBasicAuth updates a secret with basic auth
func (ct *controllerTest) UpdateSecretWithBasicAuth(username, password string) error {
	_, err := ct.k8sClient.CoreV1().Secrets(testNamespace).Update(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      authSecretName,
		},
		Data: map[string][]byte{
			v1beta1.BasicAuthUsernameKey: []byte(username),
			v1beta1.BasicAuthPasswordKey: []byte(password),
		},
	})
	return err
}

// MarkClusterServiceClassRemoved marks the cluster service class to be removed (sets the RemovedFromBrokerCatalog flag to true)
func (ct *controllerTest) MarkClusterServiceClassRemoved() error {
	csc, err := ct.scInterface.ClusterServiceClasses().Get(testClassExternalID, metav1.GetOptions{})
	if err != nil {
		return err
	}
	csc.Status.RemovedFromBrokerCatalog = true
	_, err = ct.scInterface.ClusterServiceClasses().UpdateStatus(csc)
	return err
}

// MarkClusterServicePlanRemoved marks the cluster service plan to be removed (sets the RemovedFromBrokerCatalog flag to true)
func (ct *controllerTest) MarkClusterServicePlanRemoved() error {
	csp, err := ct.scInterface.ClusterServicePlans().Get(testPlanExternalID, metav1.GetOptions{})
	if err != nil {
		return err
	}
	csp.Status.RemovedFromBrokerCatalog = true
	_, err = ct.scInterface.ClusterServicePlans().UpdateStatus(csp)
	return err
}

// AssertClusterServiceClassAndPlan verifies the both the ClusterServiceClass and ClusterServicePlan are present
func (ct *controllerTest) AssertClusterServiceClassAndPlan(t *testing.T) {
	err := ct.WaitForClusterServiceClass()
	if err != nil {
		t.Fatal(err)
	}

	err = ct.WaitForClusterServicePlan()
	if err != nil {
		t.Fatal(err)
	}
}

// AssertClusterServiceClassAndPlan verifies the both the ClusterServiceClass and ClusterServicePlan are present
func (ct *controllerTest) SetClusterServiceClassLabels() error {
	csHandler := mutation.CreateUpdateHandler{}

	csc, err := ct.scInterface.ClusterServiceClasses().Get(testClassExternalID, metav1.GetOptions{})
	if err != nil {
		return err
	}
	csHandler.SyncLabels(csc)
	_, err = ct.scInterface.ClusterServiceClasses().UpdateStatus(csc)
	return err
}

// SetCatalogReactionError sets the catalog call to always return an error
func (ct *controllerTest) SetCatalogReactionError() {
	ct.fakeOSBClient.CatalogReaction = &fakeosb.CatalogReaction{
		Error: errors.New("ooops"),
	}
}

// WaitForReadyBinding waits until the ServiceBinding is in Ready state
func (ct *controllerTest) WaitForReadyBinding() error {
	return ct.waitForBindingStatusCondition(v1beta1.ServiceBindingCondition{
		Type:   v1beta1.ServiceBindingConditionReady,
		Status: v1beta1.ConditionTrue,
	})
}

// WaitForNotReadyBinding waits until the ServiceBinding is in NotReady state
func (ct *controllerTest) WaitForNotReadyBinding() error {
	return ct.waitForBindingStatusCondition(v1beta1.ServiceBindingCondition{
		Type:   v1beta1.ServiceBindingConditionReady,
		Status: v1beta1.ConditionFalse,
	})
}

// WaitForUnbindFailed waits for the ServiceBinding to be marked as failed because of the unbind action
func (ct *controllerTest) WaitForUnbindFailed() error {
	return ct.waitForBindingStatusCondition(v1beta1.ServiceBindingCondition{
		Type:   v1beta1.ServiceBindingConditionReady,
		Status: v1beta1.ConditionUnknown,
		Reason: "UnbindCallFailed",
	})
}

// WaitForNotReadyBinding waits until the ServiceBinding is in InProgress state
func (ct *controllerTest) WaitForBindingInProgress() error {
	return ct.waitForBindingStatusCondition(v1beta1.ServiceBindingCondition{
		Type:   v1beta1.ServiceBindingConditionReady,
		Status: v1beta1.ConditionFalse,
		Reason: "Binding",
	})
}

// WaitForNotReadyBinding waits until the ServiceBinding completes the orphan mitigation
func (ct *controllerTest) WaitForBindingOrphanMitigationSuccessful() error {
	return ct.waitForBindingStatusCondition(v1beta1.ServiceBindingCondition{
		Type:   v1beta1.ServiceBindingConditionReady,
		Status: v1beta1.ConditionFalse,
		Reason: "OrphanMitigationSuccessful",
	})
}

// WaitForBindingFailed waits unit the ServiceBinding is in Failed state
func (ct *controllerTest) WaitForBindingFailed() error {
	return ct.waitForBindingStatusCondition(v1beta1.ServiceBindingCondition{
		Type:   v1beta1.ServiceBindingConditionFailed,
		Status: v1beta1.ConditionTrue,
	})
}

// WaitForUnbindStatus waits unit the ServiceBinding will have the given status
func (ct *controllerTest) WaitForUnbindStatus(status v1beta1.ServiceBindingUnbindStatus) error {
	var lastBinding *v1beta1.ServiceBinding
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		binding, err := ct.scInterface.ServiceBindings(testNamespace).Get(testBindingName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Binding: %v", err)
		}

		if binding.Status.UnbindStatus == status {
			return true, nil
		}

		lastBinding = binding
		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("binding with proper unbinding status not found, the existing binding status: %+v", lastBinding.Status)
	}
	return err
}

// WaitForDeprovisionStatus waits unit the ServiceInstance will have the given status
func (ct *controllerTest) WaitForDeprovisionStatus(status v1beta1.ServiceInstanceDeprovisionStatus) error {
	var lastInstance *v1beta1.ServiceInstance
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		si, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Binding: %v", err)
		}

		if si.Status.DeprovisionStatus == status {
			return true, nil
		}

		lastInstance = si
		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("service instance with proper deprovision status not found, "+
			"the existing service instance status: %+v", lastInstance.Status)
	}
	return err
}

// waitForBindingStatusCondition waits until ServiceBinding will have the given condition
func (ct *controllerTest) waitForBindingStatusCondition(condition v1beta1.ServiceBindingCondition) error {
	var lastBinding *v1beta1.ServiceBinding
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		binding, err := ct.scInterface.ServiceBindings(testNamespace).Get(testBindingName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Binding: %v", err)
		}

		for _, cond := range binding.Status.Conditions {
			if condition.Type == cond.Type && condition.Status == cond.Status {
				if condition.Reason == "" || condition.Reason == cond.Reason {
					return true, nil
				}
			}
		}
		lastBinding = binding
		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("binding with proper state not found, the existing binding status: %+v", lastBinding.Status)
	}
	return err
}

// WaitForServiceInstanceRemoved waits until the ServiceInstance will be removed
func (ct *controllerTest) WaitForServiceInstanceRemoved() error {
	var lastInstance *v1beta1.ServiceInstance
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		lastInstance = instance
		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("the service instance still exists: %+v", lastInstance)
	}
	return err
}

// WaitForReadyInstance waits until the ServiceInstance will be marked as ready
func (ct *controllerTest) WaitForReadyInstance() error {
	return ct.waitForInstanceCondition(v1beta1.ServiceInstanceCondition{
		Type:   v1beta1.ServiceInstanceConditionReady,
		Status: v1beta1.ConditionTrue,
	})
}

// WaitForInstanceUpdating waits until the ServiceInstance will be in update state
func (ct *controllerTest) WaitForInstanceUpdating() error {
	return ct.waitForInstanceCondition(v1beta1.ServiceInstanceCondition{
		Type:   v1beta1.ServiceInstanceConditionReady,
		Status: v1beta1.ConditionFalse,
		Reason: "UpdatingInstance",
	})
}

// WaitForServiceInstanceRemoved waits until the ServiceInstance will in given condition
func (ct *controllerTest) waitForInstanceCondition(condition v1beta1.ServiceInstanceCondition) error {
	var lastInstance *v1beta1.ServiceInstance
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Instance: %v", err)
		}
		lastInstance = instance

		for _, cond := range instance.Status.Conditions {
			if condition.Type == cond.Type && condition.Status == cond.Status {
				if condition.Reason == "" || condition.Reason == cond.Reason {
					return true, nil
				}
			}
		}
		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("the instance is in expected state (expected condition %+v), current status: %+v", condition, lastInstance.Status)
	}
	return err
}

// WaitForAsyncProvisioningInProgress waits until the ServiceInstance will be in process of async provisioning
func (ct *controllerTest) WaitForAsyncProvisioningInProgress() error {
	var lastInstance *v1beta1.ServiceInstance
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting ServiceInstance: %v", err)
		}
		lastInstance = instance

		if instance.Status.AsyncOpInProgress {
			return true, nil
		}

		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("the instance is not in progress, current status: %+v", lastInstance.Status)
	}
	return err
}

// WaitForReadyBroker waits until the ServiceBroker will be in Ready state
func (ct *controllerTest) WaitForReadyBroker() error {
	condition := v1beta1.ServiceBrokerCondition{
		Type:   v1beta1.ServiceBrokerConditionReady,
		Status: v1beta1.ConditionTrue,
	}

	var lastBroker *v1beta1.ClusterServiceBroker
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		broker, err := ct.scInterface.ClusterServiceBrokers().Get(testClusterServiceBrokerName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Broker: %v", err)
		}
		lastBroker = broker

		for _, cond := range broker.Status.Conditions {
			if condition.Type == cond.Type && condition.Status == cond.Status {
				if condition.Reason == "" || condition.Reason == cond.Reason {
					return true, nil
				}
			}
		}

		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("the broker is not ready, current status: %+v", lastBroker.Status)
	}
	return err
}

// WaitForClusterServiceClass waits until the ClusterServiceClass will be present
func (ct *controllerTest) WaitForClusterServiceClass() error {
	return wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		_, err := ct.scInterface.ClusterServiceClasses().Get(testClassExternalID, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}

		return false, err
	})
}

// WaitForClusterServiceClassToNotExists waits until the ClusterServiceClass will be removed
func (ct *controllerTest) WaitForClusterServiceClassToNotExists() error {
	return wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		_, err := ct.scInterface.ClusterServiceClasses().Get(testClassExternalID, metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}

		return false, err
	})
}

// WaitForClusterServicePlanToNotExists waits until the ClusterServicePlan will be removed
func (ct *controllerTest) WaitForClusterServicePlanToNotExists() error {
	return wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		_, err := ct.scInterface.ClusterServicePlans().Get(testPlanExternalID, metav1.GetOptions{})
		if err != nil && apierrors.IsNotFound(err) {
			return true, nil
		}

		return false, err
	})
}

// WaitForClusterServicePlan waits until the ClusterServicePlan will be present
func (ct *controllerTest) WaitForClusterServicePlan() error {
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		_, err := ct.scInterface.ClusterServicePlans().Get(testPlanExternalID, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}

		return false, err
	})
	if err == wait.ErrWaitTimeout {
		plans, e := ct.scInterface.ClusterServicePlans().List(metav1.ListOptions{})
		if e != nil {
			return err
		}
		return fmt.Errorf("plan %v not found, existing plans: %v", testPlanExternalID, plans)
	}
	return err
}

// AssertOSBRequestsUsername asserts the OriginatingIdentity request field
func (ct *controllerTest) AssertOSBRequestsUsername(t *testing.T) {
	for _, action := range ct.fakeOSBClient.Actions() {
		var oi *osb.OriginatingIdentity
		switch request := action.Request.(type) {
		case *osb.ProvisionRequest:
			oi = request.OriginatingIdentity
		case *osb.UpdateInstanceRequest:
			oi = request.OriginatingIdentity
		case *osb.DeprovisionRequest:
			oi = request.OriginatingIdentity
		case *osb.BindRequest:
			oi = request.OriginatingIdentity
		case *osb.UnbindRequest:
			oi = request.OriginatingIdentity
		case *osb.LastOperationRequest:
			oi = request.OriginatingIdentity
		default:
			continue
		}

		require.NotNil(t, oi, "originating identity of the request %v must not be nil", action.Type)

		oiValues := make(map[string]interface{})
		require.NoError(t, json.Unmarshal([]byte(oi.Value), &oiValues))

		if e, a := testUsername, oiValues["username"]; e != a {
			t.Fatalf("unexpected username in originating identity: expected %q, got %q", e, a)
		}
	}
}

// v1Now returns pointer to the current time in metav1.Time type
func (ct *controllerTest) v1Now() *metav1.Time {
	n := metav1.NewTime(time.Now())
	return &n
}

// TimeoutError simulates timeout error in provision ServiceInstance test
type TimeoutError string

// Timeout method require for TimeoutError type to meet the url/timeout interface
func (e TimeoutError) Timeout() bool {
	return true
}

// Error returns the TimeoutError as a string
func (e TimeoutError) Error() string {
	return string(e)
}

// SetupEmptyPlanListForOSBClient sets up fake OSB client response to return plans which not exist in any ServiceInstance
func (ct *controllerTest) SetupEmptyPlanListForOSBClient() {
	ct.fakeOSBClient.CatalogReaction.(*fakeosb.CatalogReaction).Response = &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:        testClusterServiceClassName,
				ID:          testClassExternalID,
				Description: "a test service",
				Bindable:    true,
				Plans: []osb.Plan{
					{
						Name:        "randomPlan",
						Free:        truePtr(),
						ID:          "randomID",
						Description: "This is plan which should not exist in any of instance",
					},
				},
			},
		},
	}
}

// WaitForInstanceCondition waits until ServiceInstance `status.conditions` field value is equal to condition in parameters
// returns error if the time limit has been reached
func (ct *controllerTest) WaitForInstanceCondition(condition v1beta1.ServiceInstanceCondition) error {
	var lastInstance *v1beta1.ServiceInstance
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Instance: %v", err)
		}

		for _, cond := range instance.Status.Conditions {
			if condition.Type == cond.Type && condition.Status == cond.Status && condition.Reason == cond.Reason {
				return true, nil
			}
		}
		lastInstance = instance
		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf(
			"instance with proper conditions not found, the existing conditions: %+v", lastInstance.Status.Conditions)
	}
	return err
}

// WaitForServiceInstanceProcessedGeneration waits until ServiceInstance parameter `Status.ObservedGeneration` is
// equal or higher than ServiceInstance `generation` value, ServiceInstance is in Ready/True status and
// ServiceInstance is not in Orphan Mitigation progress
func (ct *controllerTest) WaitForServiceInstanceProcessedGeneration(generation int64) error {
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Instance: %v", err)
		}

		if instance.Status.ObservedGeneration >= generation &&
			isServiceInstanceConditionTrue(instance) &&
			!instance.Status.OrphanMitigationInProgress {
			return true, nil
		}

		return false, nil
	})

	if err == wait.ErrWaitTimeout {
		return fmt.Errorf(
			"instance with proper ProcessedGeneration status not found")
	}
	return err
}

func isServiceInstanceConditionTrue(instance *v1beta1.ServiceInstance) bool {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == v1beta1.ServiceInstanceConditionReady || cond.Type == v1beta1.ServiceInstanceConditionFailed {
			return cond.Status == v1beta1.ConditionTrue
		}
	}

	return false
}

// AssertServiceInstanceHasNoCondition makes sure ServiceInstance is in not specific condition
func (ct *controllerTest) AssertServiceInstanceHasNoCondition(t *testing.T, cond v1beta1.ServiceInstanceCondition) {
	instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error getting Instance: %v", err)
	}

	for _, condition := range instance.Status.Conditions {
		if t1, t2 := condition.Type, cond.Type; t1 == t2 {
			if s1, s2 := condition.Status, cond.Status; s1 == s2 {
				t.Fatalf(
					"unexpected condition status: expected %v, got %v or \n "+
						"unexpected condition type: expected %v, got %v", s2, s1, t2, t1)
			}
		}
	}
}

// AssertServiceInstanceOrphanMitigationStatus makes sure ServiceInstance is/or is not in Orphan Mitigation progress
func (ct *controllerTest) AssertServiceInstanceOrphanMitigationStatus(t *testing.T, state bool) {
	instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error getting Instance: %v", err)
	}

	if om := instance.Status.OrphanMitigationInProgress; om != state {
		t.Fatalf("unexpected OrphanMitigationInProgress status: expected %v, got %v", state, om)
	}
}

// CreateClusterServiceClass creates ClusterServiceClass with default parameters
func (ct *controllerTest) CreateClusterServiceClass() error {
	serviceClass := &v1beta1.ClusterServiceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: testClassExternalID,
		},
		Spec: v1beta1.ClusterServiceClassSpec{
			ClusterServiceBrokerName: testClusterServiceBrokerName,
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalID:   testClassExternalID,
				ExternalName: testClusterServiceClassName,
				Description:  "a test service",
				Bindable:     true,
			},
		},
	}
	if _, err := ct.scInterface.ClusterServiceClasses().Create(serviceClass); err != nil {
		return err
	}

	return nil
}

// CreateClusterServicePlan creates CreateClusterServicePlan with default parameters
func (ct *controllerTest) CreateClusterServicePlan() error {
	servicePlan := &v1beta1.ClusterServicePlan{
		ObjectMeta: metav1.ObjectMeta{
			Name: testPlanExternalID,
		},
		Spec: v1beta1.ClusterServicePlanSpec{
			ClusterServiceBrokerName: testClusterServicePlanName,
		},
	}
	if _, err := ct.scInterface.ClusterServicePlans().Create(servicePlan); err != nil {
		return err
	}

	return nil
}

// UpdateServiceInstanceExternalPlanName updates ServiceInstance plan by plan ID
func (ct *controllerTest) UpdateServiceInstanceExternalPlanName(planID string) (int64, error) {
	instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("error getting Instance: %v", err)
	}

	instance.Spec.ClusterServicePlanExternalName = planID
	instance.Spec.ClusterServicePlanRef = &v1beta1.ClusterObjectReference{
		Name: planID,
	}

	instance.Generation = instance.Generation + 1
	updatedInstance, err := ct.scInterface.ServiceInstances(testNamespace).Update(instance)

	if err != nil {
		return 0, fmt.Errorf("error updating Instance: %v", err)
	}

	return updatedInstance.Generation, nil
}

// UpdateServiceInstanceInternalPlanName updates ServiceInstance plan by plan name
// CAUTION: because Plan refs are added by a Webhook tests require adds planRef before update ServiceInstance
func (ct *controllerTest) UpdateServiceInstanceInternalPlanName(planName string) (int64, error) {
	instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("error getting Instance: %v", err)
	}

	instance.Spec.ClusterServicePlanName = planName
	instance.Spec.ClusterServicePlanRef = &v1beta1.ClusterObjectReference{
		Name: testOtherPlanExternalID,
	}

	instance.Generation = instance.Generation + 1
	updatedInstance, err := ct.scInterface.ServiceInstances(testNamespace).Update(instance)

	if err != nil {
		return 0, fmt.Errorf("error updating Instance: %v", err)
	}

	return updatedInstance.Generation, nil
}

// CreateServiceInstanceWithCustomParameters creates ServiceInstance with parameters from map or
// by adding reference to Secret. If parameters are empty method creates ServiceInstance without parameters
func (ct *controllerTest) CreateServiceInstanceWithCustomParameters(withParam, paramFromSecret bool) error {
	var params map[string]interface{}
	var paramsFrom []v1beta1.ParametersFromSource

	if withParam {
		params = map[string]interface{}{
			"param-key": "param-value",
		}
	}

	if paramFromSecret {
		paramsFrom = []v1beta1.ParametersFromSource{
			{
				SecretKeyRef: &v1beta1.SecretKeyReference{
					Name: secretNameWithParameters,
					Key:  secretKeyWithParameters,
				},
			},
		}
	}

	var err error
	if withParam || paramFromSecret {
		_, err = ct.CreateServiceInstanceWithParameters(params, paramsFrom)
	} else {
		err = ct.CreateServiceInstance()
	}

	if err != nil {
		return err
	}

	return nil
}

// CreateServiceInstanceWithParameters creates ServiceInstance with parameters from map and by adding
// Secret reference
func (ct *controllerTest) CreateServiceInstanceWithParameters(
	params map[string]interface{},
	paramsFrom []v1beta1.ParametersFromSource) (*v1beta1.ServiceInstance, error) {
	rawParams, err := convertParametersIntoRawExtension(params)
	if err != nil {
		return nil, err
	}

	instance := &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testServiceInstanceName,
			Finalizers: []string{v1beta1.FinalizerServiceCatalog},
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: testClusterServiceClassName,
				ClusterServicePlanExternalName:  testClusterServicePlanName,
			},
			ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
				Name: testPlanExternalID,
			},
			ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
				Name: testClassExternalID,
			},
			ExternalID:     testExternalID,
			Parameters:     rawParams,
			ParametersFrom: paramsFrom,
		},
	}

	_, err = ct.scInterface.ServiceInstances(testNamespace).Create(instance)
	if err != nil {
		return nil, err
	}

	return instance, err
}

// UpdateCustomServiceInstanceParameters updates ServiceInstance with specific parameters. Method updates
// directly parameters, parameters by adding Secret reference, removes parameters or removes reference to Secret
func (ct *controllerTest) UpdateCustomServiceInstanceParameters(
	update, updateFromSecret, delete, deleteFromSecret bool) (int64, error) {
	instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	if update {
		instanceParam, err := convertParametersIntoRawExtension(map[string]interface{}{"param-key": "new-param-value"})
		if err != nil {
			return 0, err
		}
		instance.Spec.Parameters = instanceParam
	}

	if delete {
		instance.Spec.Parameters = nil
	}

	if updateFromSecret {
		instance.Spec.ParametersFrom = []v1beta1.ParametersFromSource{
			{
				SecretKeyRef: &v1beta1.SecretKeyReference{
					Name: otherSecretNameWithParameters,
					Key:  otherSecretKeyWithParameters,
				},
			},
		}
	}

	if deleteFromSecret {
		instance.Spec.ParametersFrom = nil
	}

	instance.Generation = instance.Generation + 1
	updatedInstance, err := ct.scInterface.ServiceInstances(testNamespace).Update(instance)
	if err != nil {
		return 0, err
	}

	return updatedInstance.Generation, nil
}

func convertParametersIntoRawExtension(parameters map[string]interface{}) (*runtime.RawExtension, error) {
	marshalledParams, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{Raw: marshalledParams}, nil
}

// CreateServiceInstanceWithInvalidParameters creates instance and updates parameters with incorrect parameters
func (ct *controllerTest) CreateServiceInstanceWithInvalidParameters() error {
	params := map[string]interface{}{
		"Name": "test-param",
		"Args": map[string]interface{}{
			"first":  "first-arg",
			"second": "second-arg",
		},
	}
	rawParams, err := convertParametersIntoRawExtension(params)
	if err != nil {
		return err
	}
	rawParams.Raw[0] = 0x21

	instance := &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testServiceInstanceName,
			Finalizers: []string{v1beta1.FinalizerServiceCatalog},
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: testClusterServiceClassName,
				ClusterServicePlanExternalName:  testClusterServicePlanName,
			},
			ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
				Name: testPlanExternalID,
			},
			ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
				Name: testClassExternalID,
			},
			ExternalID:     testExternalID,
			Parameters:     rawParams,
			ParametersFrom: nil,
		},
	}

	_, err = ct.scInterface.ServiceInstances(testNamespace).Create(instance)
	if err != nil {
		return err
	}

	return nil
}

// AssertObservedGenerationIsCorrect makes sure ServiceInstance status `ObservedGeneration` parameter is not
// equal to ServiceInstance `Generation` parameter
func (ct *controllerTest) AssertObservedGenerationIsCorrect(t *testing.T) {
	instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if g, og := instance.Generation, instance.Status.ObservedGeneration; g != og {
		t.Fatalf("latest generation not observed: generation: %v, observed: %v", g, og)
	}
}

// WaitForReadyUpdateInstance waits for ServiceInstance when Generation parameter will be equal to
// ObservedGeneration status parameter
func (ct *controllerTest) WaitForReadyUpdateInstance() error {
	err := wait.PollImmediate(pollingInterval, pollingTimeout, func() (bool, error) {
		instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("error getting Instance: %v", err)
		}

		if g, og := instance.Generation, instance.Status.ObservedGeneration; g != og {
			return true, nil
		}

		return false, nil
	})
	if err == wait.ErrWaitTimeout {
		return fmt.Errorf("ServiceInstance ObservedGeneration status parameter is out of date")
	}
	return err
}

// SetErrorReactionForProvisioningToOSBClient sets up DynamicProvisionReaction for fake osb client with specific
// error status code
func (ct *controllerTest) SetErrorReactionForProvisioningToOSBClient(statusCode int) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()

	ct.fakeOSBClient.ProvisionReaction = fakeosb.DynamicProvisionReaction(
		func(_ *osb.ProvisionRequest) (*osb.ProvisionResponse, error) {
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   statusCode,
				ErrorMessage: strPtr("error message"),
				Description:  strPtr("response description"),
			}
		})
}

// SetCustomErrorReactionForProvisioningToOSBClient sets up DynamicProvisionReaction for fake osb client
// with specific response
func (ct *controllerTest) SetCustomErrorReactionForProvisioningToOSBClient(response error) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()

	ct.fakeOSBClient.ProvisionReaction = fakeosb.DynamicProvisionReaction(
		func(_ *osb.ProvisionRequest) (*osb.ProvisionResponse, error) {
			return nil, response
		})
}

// SetErrorReactionForDeprovisioningToOSBClient sets up DynamicDeprovisionReaction for fake osb client with specific
// error status code. Method allows blocking deprovision response
func (ct *controllerTest) SetErrorReactionForDeprovisioningToOSBClient(statusCode int, block <-chan bool) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()

	blockDeprovision := true
	ct.fakeOSBClient.DeprovisionReaction = fakeosb.DynamicDeprovisionReaction(
		func(_ *osb.DeprovisionRequest) (*osb.DeprovisionResponse, error) {
			for blockDeprovision {
				blockDeprovision = <-block
			}
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   statusCode,
				ErrorMessage: strPtr("temporary deprovision error"),
			}
		})
}

// SetSuccessfullyReactionForProvisioningToOSBClient sets up DynamicProvisionReaction for fake osb client
// with success response
func (ct *controllerTest) SetSuccessfullyReactionForProvisioningToOSBClient() {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()

	ct.fakeOSBClient.ProvisionReaction = fakeosb.DynamicProvisionReaction(
		func(_ *osb.ProvisionRequest) (*osb.ProvisionResponse, error) {
			return &osb.ProvisionResponse{}, nil
		})
}

// SetSuccessfullyReactionForDeprovisioningToOSBClient sets up DynamicDeprovisionReaction for fake osb client
// with success response
func (ct *controllerTest) SetSuccessfullyReactionForDeprovisioningToOSBClient() {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()

	ct.fakeOSBClient.DeprovisionReaction = fakeosb.DynamicDeprovisionReaction(
		func(_ *osb.DeprovisionRequest) (*osb.DeprovisionResponse, error) {
			return &osb.DeprovisionResponse{}, nil
		})
}

// AssertLastOSBUpdatePlanID makes sure osb client action with type "UpdateInstance"
// contains specific plan ID in request body parameters
func (ct *controllerTest) AssertLastOSBUpdatePlanID(t *testing.T) {
	for _, planID := range ct.fetchAllPlansFromUpdateActions() {
		if planID == testOtherPlanExternalID {
			return
		}
	}

	t.Fatalf("expected ServicePlan %q not exist", testOtherPlanExternalID)
}

func (ct *controllerTest) fetchAllPlansFromUpdateActions() []string {
	var plans []string
	actions := ct.fakeOSBClient.Actions()
	for _, action := range actions {
		if action.Type == fakeosb.UpdateInstance {
			request := action.Request.(*osb.UpdateInstanceRequest)
			if request.PlanID == nil {
				continue
			}

			plans = append(plans, *request.PlanID)
		}
	}

	return plans
}

// AssertBrokerActionWithParametersExist makes sure osb client action with type "UpdateInstance"
// contains specific parameters in request body parameters
func (ct *controllerTest) AssertBrokerUpdateActionWithParametersExist(t *testing.T, parameters map[string]interface{}) {
	actions := ct.fakeOSBClient.Actions()
	for _, action := range actions {
		if action.Type != fakeosb.UpdateInstance {
			continue
		}

		request := action.Request.(*osb.UpdateInstanceRequest)
		if !reflect.DeepEqual(request.Parameters, parameters) {
			t.Fatalf("unexpected parameters: expected %v, got %v", parameters, request.Parameters)
		}
	}
}

// CreateSecretsForServiceInstanceWithSecretParams creates Secrets with specific parameters
func (ct *controllerTest) CreateSecretsForServiceInstanceWithSecretParams() error {
	_, err := ct.k8sClient.CoreV1().Secrets(testNamespace).Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      secretNameWithParameters,
		},
		Data: map[string][]byte{
			secretKeyWithParameters: []byte(`{"secret-param-key":"secret-param-value"}`),
		},
	})
	if err != nil {
		return err
	}

	_, err = ct.k8sClient.CoreV1().Secrets(testNamespace).Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      otherSecretNameWithParameters,
		},
		Data: map[string][]byte{
			otherSecretKeyWithParameters: []byte(`{"other-secret-param-key":"other-secret-param-value"}`),
		},
	})

	return err
}

// SetSimpleErrorUpdateInstanceReaction sets up DynamicUpdateInstanceReaction for fake osb client
// which returns simple error response during three first call and success response after them
func (ct *controllerTest) SetSimpleErrorUpdateInstanceReaction() {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()

	numberOfPolls := 0
	numberOfInProgressResponses := 2

	ct.fakeOSBClient.UpdateInstanceReaction = fakeosb.DynamicUpdateInstanceReaction(
		func(_ *osb.UpdateInstanceRequest) (*osb.UpdateInstanceResponse, error) {
			numberOfPolls++
			if numberOfPolls > numberOfInProgressResponses {
				return &osb.UpdateInstanceResponse{}, nil
			}
			return nil, errors.New("fake update error")
		})
}

// SetErrorUpdateInstanceReaction sets up DynamicUpdateInstanceReaction for fake osb client
// which returns specific error response during three first call and success response after them
func (ct *controllerTest) SetErrorUpdateInstanceReaction() {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()

	numberOfPolls := 0
	numberOfInProgressResponses := 2

	ct.fakeOSBClient.UpdateInstanceReaction = fakeosb.DynamicUpdateInstanceReaction(
		func(_ *osb.UpdateInstanceRequest) (*osb.UpdateInstanceResponse, error) {
			numberOfPolls++
			if numberOfPolls > numberOfInProgressResponses {
				return &osb.UpdateInstanceResponse{}, nil
			}
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusConflict,
				ErrorMessage: strPtr("OutOfQuota"),
				Description:  strPtr("You're out of quota!"),
			}
		})
}

// SetUpdateServiceInstanceResponseWithDashboardURL sets up UpdateInstanceReaction for fake osb client
// with specific url under the parameter `DashboardURL`
func (ct *controllerTest) SetUpdateServiceInstanceResponseWithDashboardURL() {
	dashURL := testDashboardURL
	ct.fakeOSBClient.UpdateInstanceReaction = &fakeosb.UpdateInstanceReaction{
		Response: &osb.UpdateInstanceResponse{
			DashboardURL: &dashURL,
		},
	}
}

// AssertServiceInstanceDashboardURL makes sure ServiceInstance `Status.DashboardURL` parameter is equal to test URL
func (ct *controllerTest) AssertServiceInstanceDashboardURL(t *testing.T) {
	instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error getting Instance: %v", err)
	}

	dashURL := testDashboardURL
	return
	if *instance.Status.DashboardURL != dashURL {
		t.Fatalf("unexpected DashboardURL: %v expected %v", instance.Status.DashboardURL, dashURL)
	}
}

// AssertServiceInstanceEmptyDashboardURL makes sure ServiceInstance `Status.DashboardURL` is empty
func (ct *controllerTest) AssertServiceInstanceEmptyDashboardURL(t *testing.T) {
	instance, err := ct.scInterface.ServiceInstances(testNamespace).Get(testServiceInstanceName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("error getting Instance: %v", err)
	}
	if instance.Status.DashboardURL != nil {
		t.Fatalf("Dashboard URL should be nil")
	}
}

// SetupFeatureGateDashboardURL sets FeatureGate behavior
func (ct *controllerTest) SetFeatureGateDashboardURL(enable bool) error {
	var format string
	if enable {
		format = "%v=true"
	} else {
		format = "%v=false"
	}

	parameter := fmt.Sprintf(format, scfeatures.UpdateDashboardURL)
	if err := utilfeature.DefaultMutableFeatureGate.Set(parameter); err != nil {
		return fmt.Errorf("Failed to enable updatable dashboard url feature: %v", err)
	}

	return nil
}

// CreateSecret creates a secret with given name and stored data
func (ct *controllerTest) CreateSecret(name string, data map[string][]byte) error {
	_, err := ct.k8sClient.CoreV1().Secrets(testNamespace).Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      name,
		},
		Data: data,
	})
	return err
}

// AssertLastBindRequest makes sure parameters are equal to parameters from last binding request
func (ct *controllerTest) AssertLastBindRequest(t *testing.T, expectedParams map[string]interface{}) {
	actions := ct.fakeOSBClient.Actions()
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		if action.Type == fakeosb.Bind {
			bindReq := action.Request.(*osb.BindRequest)
			assert.Equal(t, expectedParams, bindReq.Parameters)
			return
		}
	}
}

// CreateServiceInstance creates a ServiceInstance which is used in testing scenarios.
func (ct *controllerTest) CreateServiceInstanceWithNonbindablePlan() error {
	_, err := ct.scInterface.ServiceInstances(testNamespace).Create(&v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: testServiceInstanceName,
			// added by a Webhook, which is not tested here
			Finalizers: []string{v1beta1.FinalizerServiceCatalog},
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: testClassExternalID,
				ClusterServicePlanExternalName:  testNonbindablePlanExternalID,
			},
			ExternalID: testExternalID,
			// Plan and Class refs are added by a Webhook, which is not tested here
			ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
				Name: testNonbindablePlanExternalID,
			},
			ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
				Name: testClassExternalID,
			},
		},
	})
	return err
}

// SetOSBProvisionReactionWithHTTPError configures the broker Provision call response as HTTPStatusCodeError
func (ct *controllerTest) SetOSBProvisionReactionWithHTTPError(code int) {
	ct.fakeOSBClient.Lock()
	defer ct.fakeOSBClient.Unlock()
	ct.fakeOSBClient.ProvisionReaction = &fakeosb.ProvisionReaction{
		Error: osb.HTTPStatusCodeError{
			StatusCode: code,
		},
	}
}

func (ct *controllerTest) CreateBindingWithParams(params map[string]interface{}, paramsFrom []v1beta1.ParametersFromSource) error {
	var parameters *runtime.RawExtension
	if params != nil {
		marshaledParams, err := json.Marshal(params)
		if err != nil {
			return err
		}
		parameters = &runtime.RawExtension{Raw: marshaledParams}
	}
	_, err := ct.scInterface.ServiceBindings(testNamespace).Create(&v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  testNamespace,
			Name:       testBindingName,
			Generation: 1,
			Finalizers: []string{v1beta1.FinalizerServiceCatalog}, // set by the Webhook
		},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: v1beta1.LocalObjectReference{
				Name: testServiceInstanceName,
			},
			ExternalID:     testServiceBindingGUID,
			SecretName:     testBindingName, // set by the webhook
			Parameters:     parameters,
			ParametersFrom: paramsFrom,
		},
	})
	return err
}

// AssertBindingData verifies the secret created by the binding - checks stored secret data.
func (ct *controllerTest) AssertBindingData(t *testing.T, expectedData map[string][]byte) {
	s, err := ct.k8sClient.CoreV1().Secrets(testNamespace).Get(testBindingName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, expectedData, s.Data)
}

func (ct *controllerTest) CreateBindingWithTransforms(transforms []v1beta1.SecretTransform) error {
	_, err := ct.scInterface.ServiceBindings(testNamespace).Create(&v1beta1.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  testNamespace,
			Name:       testBindingName,
			Generation: 1,
			Finalizers: []string{v1beta1.FinalizerServiceCatalog}, // set by the Webhook
		},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: v1beta1.LocalObjectReference{
				Name: testServiceInstanceName,
			},
			ExternalID:       testServiceBindingGUID,
			SecretName:       testBindingName, // set by the webhook
			SecretTransforms: transforms,
		},
	})
	return err
}

// fixtureHappyPathBrokerClientConfig returns fake configuration for OSB client used in testing scenario
func fixtureHappyPathBrokerClientConfig() fakeosb.FakeClientConfiguration {
	return fakeosb.FakeClientConfiguration{
		CatalogReaction: &fakeosb.CatalogReaction{
			Response: fixtureCatalogResponse(),
		},
		ProvisionReaction: &fakeosb.ProvisionReaction{
			Response: &osb.ProvisionResponse{},
		},
		UpdateInstanceReaction: &fakeosb.UpdateInstanceReaction{
			Response: &osb.UpdateInstanceResponse{},
		},
		DeprovisionReaction: &fakeosb.DeprovisionReaction{
			Response: &osb.DeprovisionResponse{},
		},
		BindReaction: &fakeosb.BindReaction{
			Response: &osb.BindResponse{
				Credentials: fixtureBindCredentials(),
			},
		},
		UnbindReaction: &fakeosb.UnbindReaction{
			Response: &osb.UnbindResponse{},
		},
		PollLastOperationReaction: &fakeosb.PollLastOperationReaction{
			Response: &osb.LastOperationResponse{
				State: osb.StateSucceeded,
			},
		},
		PollBindingLastOperationReaction: &fakeosb.PollBindingLastOperationReaction{
			Response: &osb.LastOperationResponse{
				State: osb.StateSucceeded,
			},
		},
		GetBindingReaction: &fakeosb.GetBindingReaction{
			Response: &osb.GetBindingResponse{
				Credentials: fixtureBindCredentials(),
			},
		},
	}
}

// fixtureCatalogResponse returns a sample response to a get catalog request.
func fixtureCatalogResponse() *osb.CatalogResponse {
	return &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          testClusterServiceClassName,
				ID:            testClassExternalID,
				Description:   "a test service",
				Bindable:      true,
				PlanUpdatable: truePtr(),
				Plans: []osb.Plan{
					{
						Name:        testClusterServicePlanName,
						Free:        truePtr(),
						ID:          testPlanExternalID,
						Description: "a test plan",
					},
					{
						Name:        testNonbindableClusterServicePlanName,
						Free:        truePtr(),
						ID:          testNonbindablePlanExternalID,
						Description: "an non-bindable test plan",
						Bindable:    falsePtr(),
					},
					{
						Name:        testOtherClusterServicePlanName,
						Free:        truePtr(),
						ID:          testOtherPlanExternalID,
						Description: "an other test plan",
					},
				},
			},
		},
	}
}

// fixtureBindCredentials returns binding credentials to include in the response
// to a bind request.
func fixtureBindCredentials() map[string]interface{} {
	return map[string]interface{}{
		"foo": "bar",
		"baz": "zap",
	}
}

func fixtureUserInfo() *v1beta1.UserInfo {
	return &v1beta1.UserInfo{
		Username: testUsername,
	}
}

func truePtr() *bool {
	b := true
	return &b
}

func falsePtr() *bool {
	b := false
	return &b
}

func strPtr(s string) *string {
	return &s
}
