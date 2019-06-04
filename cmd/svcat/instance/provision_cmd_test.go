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

package instance_test

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/command"
	. "github.com/kubernetes-sigs/service-catalog/cmd/svcat/instance"
	"github.com/kubernetes-sigs/service-catalog/cmd/svcat/test"
	_ "github.com/kubernetes-sigs/service-catalog/internal/test"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat"
	servicecatalog "github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog"
	"github.com/kubernetes-sigs/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Provision Command", func() {
	Describe("NewProvisionCmd", func() {
		It("Builds and returns a cobra command with the correct flags", func() {
			cxt := &command.Context{}
			cmd := NewProvisionCmd(cxt)

			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("provision NAME --plan PLAN --class CLASS"))
			Expect(cmd.Short).To(ContainSubstring("Create a new instance of a service"))
			Expect(cmd.Example).To(ContainSubstring("svcat provision wordpress-mysql-instance --class mysqldb --plan free"))
			Expect(len(cmd.Aliases)).To(Equal(0))

			flag := cmd.Flags().Lookup("plan")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Usage).To(ContainSubstring("The plan name (Required)"))

			flag = cmd.Flags().Lookup("class")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Usage).To(ContainSubstring("The class name (Required)"))

			flag = cmd.Flags().Lookup("external-id")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Usage).To(ContainSubstring("The ID of the instance for use with the OSB SB API (Optional)"))

			flag = cmd.Flags().Lookup("kube-name")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Usage).To(ContainSubstring("Whether or not to interpret the Class/Plan names as Kubernetes names (the default is by external name)"))

			flag = cmd.Flags().Lookup("param")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Usage).To(ContainSubstring("Additional parameter to use when provisioning the service, format: NAME=VALUE. Cannot be combined with --params-json, Sensitive information should be placed in a secret and specified with --secret"))

			flag = cmd.Flags().Lookup("secret")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Usage).To(ContainSubstring("Additional parameter, whose value is stored in a secret, to use when provisioning the service, format: SECRET[KEY]"))

			flag = cmd.Flags().Lookup("wait")
			Expect(flag).NotTo(BeNil())
			flag = cmd.Flags().Lookup("namespace")
			Expect(flag).NotTo(BeNil())
		})
	})
	Describe("Validate", func() {
		It("succeeds if an instance name is provided", func() {
			cmd := ProvisionCmd{}
			err := cmd.Validate([]string{"bananainstance"})
			Expect(err).NotTo(HaveOccurred())
		})
		It("errors if no instance name is provided", func() {
			cmd := ProvisionCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("an instance name is required"))
		})
		It("errors if both json params and raw params are provided", func() {
			cmd := ProvisionCmd{
				JSONParams: "{\"foo\":\"bar\"}",
				RawParams:  []string{"a=b"},
			}
			err := cmd.Validate([]string{"bananainstance"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("--params-json cannot be used with --param"))
		})
		It("succeeds only if the provided json params are parseable json", func() {
			cmd := ProvisionCmd{
				JSONParams: "{\"foo\":\"bar\"}",
			}
			err := cmd.Validate([]string{"bananainstance"})
			Expect(err).NotTo(HaveOccurred())
			p := make(map[string]interface{})
			p["foo"] = "bar"
			Expect(cmd.Params).To(Equal(p))
		})
		It("successfully parses raw params into the params map", func() {
			cmd := ProvisionCmd{
				RawParams: []string{"a=b"},
			}
			err := cmd.Validate([]string{"bananainstance"})
			Expect(err).NotTo(HaveOccurred())
			p := make(map[string]interface{})
			p["a"] = "b"
			Expect(cmd.Params).To(Equal(p))
		})
		It("errors if the provided json params are not parseable", func() {
			cmd := ProvisionCmd{
				JSONParams: "foo=bar",
			}
			err := cmd.Validate([]string{"bananainstance"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid --params-json value (invalid parameters (foo=bar))"))
		})
		It("parses secrets into the secrets map", func() {
			cmd := ProvisionCmd{
				RawSecrets: []string{"foo[bar]"},
			}
			err := cmd.Validate([]string{"bananainstance"})
			Expect(err).NotTo(HaveOccurred())
			s := make(map[string]string)
			s["foo"] = "bar"
			Expect(cmd.Secrets).To(Equal(s))
		})
		It("errors if secrets aren't parseable", func() {
			cmd := ProvisionCmd{
				RawSecrets: []string{"foo=bar"},
			}
			err := cmd.Validate([]string{"bananainstance"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid --secret value (invalid parameter (foo=bar), must be in MAP[KEY] format)"))
		})
	})
	Describe("Run", func() {
		var (
			classKubeName    string
			className        string
			classToReturn    servicecatalog.Class
			cxt              *command.Context
			externalID       string
			fakeApp          *svcat.App
			fakeSDK          *servicecatalogfakes.FakeSvcatClient
			instanceName     string
			instanceToReturn *v1beta1.ServiceInstance
			namespace        string
			outputBuffer     *bytes.Buffer
			params           map[string]interface{}
			planKubeName     string
			planName         string
			planToReturn     servicecatalog.Plan
			secrets          map[string]string
		)
		BeforeEach(func() {
			classKubeName = "mysqlclass1234"
			className = "mysqlclass"
			externalID = "mysqlexternalid"
			instanceName = "myMysql"
			namespace = "foobarnamespace"
			params = make(map[string]interface{})
			params["foo"] = "bar"
			planKubeName = "mysqlplan1234"
			planName = "10mb"
			secrets = make(map[string]string)
			secrets["foo"] = "bar"
			paramsJSON, err := json.Marshal(params)
			Expect(err).To(BeNil())
			specParams := &runtime.RawExtension{Raw: paramsJSON}
			instanceToReturn = &v1beta1.ServiceInstance{
				ObjectMeta: v1.ObjectMeta{
					Name:      instanceName,
					Namespace: namespace,
				},
				Spec: v1beta1.ServiceInstanceSpec{
					ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
						Name: classKubeName,
					},
					ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
						Name: planKubeName,
					},
					ExternalID: externalID,
					Parameters: specParams,
					PlanReference: v1beta1.PlanReference{
						ClusterServiceClassExternalName: className,
						ClusterServicePlanExternalName:  planName,
					},
				},
			}
			classToReturn = &v1beta1.ClusterServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name: classKubeName,
				},
			}
			planToReturn = &v1beta1.ClusterServicePlan{
				ObjectMeta: v1.ObjectMeta{
					Name: planKubeName,
				},
			}

			fakeSDK = new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.ProvisionReturns(instanceToReturn, nil)
			fakeSDK.RetrieveClassByNameReturns(classToReturn, nil)
			fakeSDK.RetrievePlanByClassIDAndNameReturns(planToReturn, nil)
			fakeApp, _ = svcat.NewApp(nil, nil, namespace)
			fakeApp.SvcatClient = fakeSDK
			outputBuffer = &bytes.Buffer{}
			cxt = svcattest.NewContext(outputBuffer, fakeApp)
		})

		It("Calls the pkg/svcat lib methods to find the correct k8s names, and then calls provision method with those names, and prints output to the user", func() {
			cmd := ProvisionCmd{
				ClassName:    className,
				ExternalID:   externalID,
				InstanceName: instanceName,
				Params:       params,
				PlanName:     planName,
				Secrets:      secrets,
				Namespaced:   command.NewNamespaced(cxt),
				Waitable:     command.NewWaitable(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Waitable.ApplyWaitFlags()

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeSDK.RetrieveClassByIDCallCount()).To(Equal(0))

			Expect(fakeSDK.RetrieveClassByNameCallCount()).To(Equal(1))
			returnedClassName, returnedScopeOpts := fakeSDK.RetrieveClassByNameArgsForCall(0)
			Expect(returnedClassName).To(Equal(className))
			Expect(returnedScopeOpts).To(Equal(servicecatalog.ScopeOptions{
				Namespace: namespace,
				Scope:     servicecatalog.AllScope,
			}))

			Expect(fakeSDK.RetrievePlanByClassIDAndNameCallCount()).To(Equal(1))
			returnedClassKubeName, returnedPlanName, returnedScopeOpts := fakeSDK.RetrievePlanByClassIDAndNameArgsForCall(0)
			Expect(returnedClassKubeName).To(Equal(classKubeName))
			Expect(returnedPlanName).To(Equal(planName))
			Expect(returnedScopeOpts).To(Equal(servicecatalog.ScopeOptions{
				Namespace: namespace,
				Scope:     servicecatalog.ClusterScope,
			}))

			Expect(fakeSDK.ProvisionCallCount()).To(Equal(1))
			returnedInstanceName, returnedClassKubeName, returnedPlanKubeName, returnedProvisionClusterInstance, returnedOpts := fakeSDK.ProvisionArgsForCall(0)
			Expect(returnedInstanceName).To(Equal(instanceName))
			Expect(returnedClassKubeName).To(Equal(classKubeName))
			Expect(returnedPlanKubeName).To(Equal(planKubeName))
			Expect(returnedProvisionClusterInstance).To(BeTrue())
			Expect(returnedOpts).NotTo(BeNil())
			opts := servicecatalog.ProvisionOptions{
				ExternalID: externalID,
				Namespace:  namespace,
				Params:     params,
				Secrets:    secrets,
			}
			Expect(*returnedOpts).To(Equal(opts))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(instanceName))
			Expect(output).To(ContainSubstring(namespace))
			Expect(output).To(ContainSubstring(className))
		})
		It("Calls the SDK's WaitForInstance method with the passed in interval and timeout when Wait==true", func() {
			interval := 1 * time.Second
			timeout := 1 * time.Minute
			fakeSDK.WaitForInstanceReturns(instanceToReturn, nil)
			cmd := ProvisionCmd{
				ClassName:    className,
				ExternalID:   externalID,
				InstanceName: instanceName,
				Params:       params,
				PlanName:     planName,
				Secrets:      secrets,
				Namespaced:   command.NewNamespaced(cxt),
				Waitable:     command.NewWaitable(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Waitable.ApplyWaitFlags()
			cmd.Wait = true
			cmd.Interval = interval
			cmd.Timeout = &timeout

			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.ProvisionCallCount()).To(Equal(1))
			returnedInstanceName, returnedClassKubeName, returnedPlanKubeName, returnedProvisionClusterInstance, returnedOpts := fakeSDK.ProvisionArgsForCall(0)
			Expect(returnedInstanceName).To(Equal(instanceName))
			Expect(returnedClassKubeName).To(Equal(classKubeName))
			Expect(returnedPlanKubeName).To(Equal(planKubeName))
			Expect(returnedProvisionClusterInstance).To(BeTrue())
			Expect(returnedOpts).NotTo(BeNil())
			opts := servicecatalog.ProvisionOptions{
				ExternalID: externalID,
				Namespace:  namespace,
				Params:     params,
				Secrets:    secrets,
			}
			Expect(*returnedOpts).To(Equal(opts))

			Expect(fakeSDK.WaitForInstanceCallCount()).To(Equal(1))
			waitNamespace, waitName, waitInterval, waitTimeout := fakeSDK.WaitForInstanceArgsForCall(0)
			Expect(waitNamespace).To(Equal(namespace))
			Expect(waitName).To(Equal(instanceName))
			Expect(waitInterval).To(Equal(interval))
			Expect(*waitTimeout).To(Equal(timeout))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring("Waiting for the instance"))
			Expect(output).To(ContainSubstring(instanceName))
			Expect(output).To(ContainSubstring(namespace))
			Expect(output).To(ContainSubstring(className))
		})
		It("sets ProvisionClusterInstance to true if provisioning a cluster class instance", func() {
			cmd := ProvisionCmd{
				ClassName:    className,
				ExternalID:   externalID,
				InstanceName: instanceName,
				Params:       params,
				PlanName:     planName,
				Secrets:      secrets,
				Namespaced:   command.NewNamespaced(cxt),
				Waitable:     command.NewWaitable(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Waitable.ApplyWaitFlags()

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(cmd.ProvisionClusterInstance).To(BeTrue())
			Expect(fakeSDK.ProvisionCallCount()).To(Equal(1))
			_, _, _, returnedProvisionClusterInstance, _ := fakeSDK.ProvisionArgsForCall(0)
			Expect(returnedProvisionClusterInstance).To(BeTrue())
		})
		It("sets scope to namespaced for RetrievePlanByClassIDAndName and sets ProvisionClusterInstance to false if provisioning a namespace class instance", func() {
			instanceToReturn = &v1beta1.ServiceInstance{
				ObjectMeta: v1.ObjectMeta{
					Name:      instanceName,
					Namespace: namespace,
				},
				Spec: v1beta1.ServiceInstanceSpec{
					ServiceClassRef: &v1beta1.LocalObjectReference{
						Name: classKubeName,
					},
					ServicePlanRef: &v1beta1.LocalObjectReference{
						Name: planKubeName,
					},
					ExternalID: externalID,
					PlanReference: v1beta1.PlanReference{
						ServiceClassExternalName: className,
						ServicePlanExternalName:  planName,
					},
				},
			}
			classToReturn = &v1beta1.ServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name: classKubeName,
				},
			}
			planToReturn = &v1beta1.ServicePlan{
				ObjectMeta: v1.ObjectMeta{
					Name: planKubeName,
				},
			}

			fakeSDK.ProvisionReturns(instanceToReturn, nil)
			fakeSDK.RetrieveClassByNameReturns(classToReturn, nil)
			fakeSDK.RetrievePlanByClassIDAndNameReturns(planToReturn, nil)

			cmd := ProvisionCmd{
				ClassName:    className,
				ExternalID:   externalID,
				InstanceName: instanceName,
				Params:       params,
				PlanName:     planName,
				Secrets:      secrets,
				Namespaced:   command.NewNamespaced(cxt),
				Waitable:     command.NewWaitable(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Waitable.ApplyWaitFlags()

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.RetrieveClassByNameCallCount()).To(Equal(1))
			returnedClassName, returnedScopeOpts := fakeSDK.RetrieveClassByNameArgsForCall(0)
			Expect(returnedClassName).To(Equal(className))
			Expect(returnedScopeOpts).To(Equal(servicecatalog.ScopeOptions{
				Namespace: namespace,
				Scope:     servicecatalog.AllScope,
			}))

			Expect(fakeSDK.RetrievePlanByClassIDAndNameCallCount()).To(Equal(1))
			returnedClassKubeName, returnedPlanName, returnedScopeOpts := fakeSDK.RetrievePlanByClassIDAndNameArgsForCall(0)
			Expect(returnedClassKubeName).To(Equal(classKubeName))
			Expect(returnedPlanName).To(Equal(planName))
			Expect(returnedScopeOpts).To(Equal(servicecatalog.ScopeOptions{
				Namespace: namespace,
				Scope:     servicecatalog.NamespaceScope,
			}))

			Expect(cmd.ProvisionClusterInstance).To(BeFalse())
			Expect(fakeSDK.ProvisionCallCount()).To(Equal(1))
			_, _, _, returnedProvisionClusterInstance, _ := fakeSDK.ProvisionArgsForCall(0)
			Expect(returnedProvisionClusterInstance).To(BeFalse())
		})
	})
})
