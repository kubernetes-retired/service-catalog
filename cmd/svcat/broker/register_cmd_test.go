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

	. "github.com/kubernetes-incubator/service-catalog/cmd/svcat/broker"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Register Command", func() {
	Describe("NewRegisterCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewRegisterCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("register NAME --url URL"))
			Expect(cmd.Short).To(ContainSubstring("Registers a new broker with service catalog"))
			Expect(cmd.Example).To(ContainSubstring("svcat register mysqlbroker --url http://mysqlbroker.com"))
			Expect(len(cmd.Aliases)).To(Equal(0))
		})
	})
	Describe("Validate", func() {
		It("errors if a broker name is not provided", func() {
			cmd := RegisterCmd{
				BrokerName: "",
				Context:    nil,
				URL:        "http://bananabroker.com",
			}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Register", func() {
		It("Calls the pkg/svcat libs Register method with the passed in variables and prints output to the user", func() {
			brokerName := "foobarbroker"
			brokerURL := "http://foobar.com"

			brokerToReturn := &v1beta1.ClusterServiceBroker{
				ObjectMeta: v1.ObjectMeta{
					Name: brokerName,
				},
				Spec: v1beta1.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
						URL: brokerURL,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.RegisterReturns(brokerToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := RegisterCmd{
				Context:    svcattest.NewContext(outputBuffer, fakeApp),
				BrokerName: brokerName,
				URL:        brokerURL,
			}
			err := cmd.Register()

			Expect(err).NotTo(HaveOccurred())
			returnedName, returnedURL := fakeSDK.RegisterArgsForCall(0)
			Expect(returnedName).To(Equal(brokerName))
			Expect(returnedURL).To(Equal(brokerURL))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(brokerName))
			Expect(output).To(ContainSubstring(brokerURL))
		})
	})
})
