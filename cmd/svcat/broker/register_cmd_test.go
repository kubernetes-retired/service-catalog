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

package broker_test

import (
	"bytes"
	"time"

	. "github.com/kubernetes-incubator/service-catalog/cmd/svcat/broker"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Register Command", func() {
	Describe("NewRegisterCmd", func() {
		It("Builds and returns a cobra command with the correct flags", func() {
			cxt := &command.Context{}
			cmd := NewRegisterCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("register NAME --url URL"))
			Expect(cmd.Short).To(ContainSubstring("Registers a new broker with service catalog"))
			Expect(cmd.Example).To(ContainSubstring("svcat register mysqlbroker --url http://mysqlbroker.com"))
			Expect(len(cmd.Aliases)).To(Equal(0))

			urlFlag := cmd.Flags().Lookup("url")
			Expect(urlFlag).NotTo(BeNil())
			Expect(urlFlag.Usage).To(ContainSubstring("The broker URL (Required)"))

			basicSecretFlag := cmd.Flags().Lookup("basic-secret")
			Expect(basicSecretFlag).NotTo(BeNil())
			Expect(basicSecretFlag.Usage).To(ContainSubstring("A secret containing basic auth (username/password) information to connect to the broker"))

			bearerSecretFlag := cmd.Flags().Lookup("bearer-secret")
			Expect(bearerSecretFlag).NotTo(BeNil())
			Expect(bearerSecretFlag.Usage).To(ContainSubstring("A secret containing a bearer token to connect to the broker"))

			caFlag := cmd.Flags().Lookup("ca")
			Expect(caFlag).NotTo(BeNil())
			Expect(caFlag.Usage).To(ContainSubstring("A file containing the CA certificate to connect to the broker"))

			classRestrictionFlag := cmd.Flags().Lookup("class-restrictions")
			Expect(classRestrictionFlag).NotTo(BeNil())
			Expect(classRestrictionFlag.Usage).To(ContainSubstring("A list of restrictions to apply to the classes allowed from the broker"))

			planRestrictionFlag := cmd.Flags().Lookup("plan-restrictions")
			Expect(planRestrictionFlag).NotTo(BeNil())
			Expect(planRestrictionFlag.Usage).To(ContainSubstring("A list of restrictions to apply to the plans allowed from the broker"))

			relistBehaviorFlag := cmd.Flags().Lookup("relist-behavior")
			Expect(relistBehaviorFlag).NotTo(BeNil())
			Expect(relistBehaviorFlag.Usage).To(ContainSubstring("Behavior for relisting the broker's catalog. Valid options are manual or duration. Defaults to duration with an interval of 15m."))

			relistDurationFlag := cmd.Flags().Lookup("relist-duration")
			Expect(relistDurationFlag).NotTo(BeNil())
			Expect(relistDurationFlag.Usage).To(ContainSubstring("Interval to refetch broker catalog when relist-behavior is set to duration, specified in human readable format: 30s, 1m, 1h"))

			skipTLSFlag := cmd.Flags().Lookup("skip-tls")
			Expect(skipTLSFlag).NotTo(BeNil())
			Expect(skipTLSFlag.Usage).To(ContainSubstring("Disables TLS certificate verification when communicating with this broker. This is strongly discouraged. You should use --ca instead."))

			waitFlag := cmd.Flags().Lookup("wait")
			Expect(waitFlag).NotTo(BeNil())
			timeoutFlag := cmd.Flags().Lookup("timeout")
			Expect(timeoutFlag).NotTo(BeNil())
			intervalFlag := cmd.Flags().Lookup("interval")
			Expect(intervalFlag).NotTo(BeNil())
		})
	})

	Describe("Validate", func() {
		It("succeeds if a broker name and url are provided", func() {
			cmd := RegisterCmd{}
			err := cmd.Validate([]string{"bananabroker", "http://bananabroker.com"})
			Expect(err).NotTo(HaveOccurred())
		})
		It("errors if a broker name is not provided", func() {
			cmd := RegisterCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("a broker name is required"))
		})
		It("errors if both basic-secret and bearer-secret are provided", func() {
			basicSecret := "basicsecret"
			bearerSecret := "bearersecret"
			cmd := RegisterCmd{
				BasicSecret:  basicSecret,
				BearerSecret: bearerSecret,
			}
			err := cmd.Validate([]string{"bananabroker", "http://bananabroker.com", "--basic-secret", basicSecret, "--bearer-secret", bearerSecret})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot use both basic auth and bearer auth"))
		})
		It("errors if a provided CA file does not exist", func() {
			cmd := RegisterCmd{
				CAFile: "/not/a/real/file",
			}
			err := cmd.Validate([]string{"bananabroker", "http://bananabroker.com"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error finding CA file"))
		})
		It("only allows valid values for relist behavior", func() {
			cmd := RegisterCmd{
				RelistBehavior: "foobar",
			}
			err := cmd.Validate([]string{"bananabroker", "http://bananabroker.com"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid --relist-duration value, allowed values are: duration, manual"))

			cmd = RegisterCmd{
				RelistBehavior: "Duration",
			}
			err = cmd.Validate([]string{"bananabroker", "http://bananabroker.com"})
			Expect(err).NotTo(HaveOccurred())

			cmd = RegisterCmd{
				RelistBehavior: "MANUAL",
			}
			err = cmd.Validate([]string{"bananabroker", "http://bananabroker.com"})
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Describe("Run", func() {
		var (
			basicSecret         string
			brokerName          string
			brokerURL           string
			certFile            string
			classRestrictions   []string
			brokerToReturn      *v1beta1.ClusterServiceBroker
			namespace           string
			metaRelistDuration  *metav1.Duration
			planRestrictions    []string
			relistBehavior      string
			relistBehaviorConst v1beta1.ServiceBrokerRelistBehavior
			relistDuration      time.Duration
			skipTLS             bool
		)
		BeforeEach(func() {
			brokerName = "foobarbroker"
			brokerURL = "http://foobar.com"
			basicSecret = "foobarsecret"
			certFile = "register_cmd_test.go"
			certFileContents := []byte("foobarCA")
			classRestrictions = []string{"foobarclassa", "foobarclassb"}
			namespace = "foobarnamespace"
			planRestrictions = []string{"foobarplana", "foobarplanb"}
			skipTLS = true
			relistBehavior = "duration"
			relistBehaviorConst = v1beta1.ServiceBrokerRelistBehaviorDuration
			relistDuration = 10 * time.Minute
			metaRelistDuration = &metav1.Duration{Duration: relistDuration}

			brokerToReturn = &v1beta1.ClusterServiceBroker{
				ObjectMeta: v1.ObjectMeta{
					Name: brokerName,
				},
				Spec: v1beta1.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
						CABundle:              certFileContents,
						InsecureSkipTLSVerify: skipTLS,
						RelistBehavior:        relistBehaviorConst,
						RelistDuration:        metaRelistDuration,
						URL:                   brokerURL,
						CatalogRestrictions: &v1beta1.CatalogRestrictions{
							ServiceClass: classRestrictions,
							ServicePlan:  planRestrictions,
						},
					},
					AuthInfo: &v1beta1.ClusterServiceBrokerAuthInfo{
						Basic: &v1beta1.ClusterBasicAuthConfig{
							SecretRef: &v1beta1.ObjectReference{
								Name:      basicSecret,
								Namespace: namespace,
							},
						},
					},
				},
			}
		})

		It("Calls the pkg/svcat libs Register method with the passed in variables and prints output to the user", func() {
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RegisterReturns(brokerToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := RegisterCmd{
				BasicSecret:       basicSecret,
				BrokerName:        brokerName,
				CAFile:            certFile,
				ClassRestrictions: classRestrictions,
				Namespaced:        command.NewNamespaced(cxt),
				PlanRestrictions:  planRestrictions,
				RelistBehavior:    relistBehavior,
				RelistDuration:    relistDuration,
				Scoped:            command.NewScoped(),
				SkipTLS:           skipTLS,
				URL:               brokerURL,
				Waitable:          command.NewWaitable(),
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Waitable.ApplyWaitFlags()
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.RegisterCallCount()).To(Equal(1))
			returnedName, returnedURL, returnedOpts, returnedScopeOpts := fakeSDK.RegisterArgsForCall(0)
			Expect(returnedName).To(Equal(brokerName))
			Expect(returnedURL).To(Equal(brokerURL))
			opts := servicecatalog.RegisterOptions{
				BasicSecret:       basicSecret,
				CAFile:            certFile,
				ClassRestrictions: classRestrictions,
				Namespace:         namespace,
				PlanRestrictions:  planRestrictions,
				RelistBehavior:    relistBehaviorConst,
				RelistDuration:    metaRelistDuration,
				SkipTLS:           skipTLS,
			}
			Expect(*returnedOpts).To(Equal(opts))
			Expect(returnedScopeOpts.Namespace).To(Equal(namespace))
			Expect(returnedScopeOpts.Scope.Matches(servicecatalog.NamespaceScope)).To(BeTrue())
			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(brokerName))
			Expect(output).To(ContainSubstring(brokerURL))
		})
		It("Passes in the bearer secret", func() {
			bearerSecret := "foobarsecret"
			brokerToReturn.Spec.AuthInfo.Basic = nil
			brokerToReturn.Spec.AuthInfo.Bearer = &v1beta1.ClusterBearerTokenAuthConfig{
				SecretRef: &v1beta1.ObjectReference{
					Name: bearerSecret,
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RegisterReturns(brokerToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := RegisterCmd{
				BearerSecret: bearerSecret,
				BrokerName:   brokerName,
				Namespaced:   command.NewNamespaced(cxt),
				Scoped:       command.NewScoped(),
				Waitable:     command.NewWaitable(),
				URL:          brokerURL,
			}
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Waitable.ApplyWaitFlags()
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.RegisterCallCount()).To(Equal(1))
			returnedName, returnedURL, returnedOpts, _ := fakeSDK.RegisterArgsForCall(0)
			Expect(returnedName).To(Equal(brokerName))
			Expect(returnedURL).To(Equal(brokerURL))
			opts := servicecatalog.RegisterOptions{
				Namespace:    namespace,
				BearerSecret: bearerSecret,
			}
			Expect(*returnedOpts).To(Equal(opts))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(brokerName))
			Expect(output).To(ContainSubstring(brokerURL))
		})
		It("Calls the SDK's WaitForBroker method with the passed in interval and timeout when Wait==true", func() {
			interval := 1 * time.Second
			timeout := 1 * time.Minute

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, namespace)
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RegisterReturns(brokerToReturn, nil)
			fakeSDK.WaitForBrokerReturns(brokerToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cxt := svcattest.NewContext(outputBuffer, fakeApp)
			cmd := RegisterCmd{
				BrokerName: brokerName,
				Namespaced: command.NewNamespaced(cxt),
				Scoped:     command.NewScoped(),
				Waitable:   command.NewWaitable(),
				URL:        brokerURL,
			}
			cmd.Wait = true
			cmd.Namespaced.ApplyNamespaceFlags(&pflag.FlagSet{})
			cmd.Waitable.ApplyWaitFlags()
			cmd.Interval = interval
			cmd.Timeout = &timeout

			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSDK.RegisterCallCount()).To(Equal(1))
			returnedName, returnedURL, returnedOpts, _ := fakeSDK.RegisterArgsForCall(0)
			Expect(returnedName).To(Equal(brokerName))
			Expect(returnedURL).To(Equal(brokerURL))
			opts := servicecatalog.RegisterOptions{
				Namespace: namespace,
			}
			Expect(*returnedOpts).To(Equal(opts))

			Expect(fakeSDK.WaitForBrokerCallCount()).To(Equal(1))
			waitName, waitInterval, waitTimeout := fakeSDK.WaitForBrokerArgsForCall(0)
			Expect(waitName).To(Equal(brokerName))
			Expect(waitInterval).To(Equal(interval))
			Expect(*waitTimeout).To(Equal(timeout))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(brokerName))
			Expect(output).To(ContainSubstring(brokerURL))
		})
	})
})
