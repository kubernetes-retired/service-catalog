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
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deregister Command", func() {
	Describe("NewDeregisterCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewDeregisterCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("deregister NAME"))
			Expect(cmd.Short).To(ContainSubstring("Deregisters an existing broker with service catalog"))
			Expect(cmd.Example).To(ContainSubstring("svcat deregister mysqlbroker"))
			Expect(len(cmd.Aliases)).To(Equal(0))
		})
	})
	Describe("Validate", func() {
		It("errors if a broker name is not provided", func() {
			cmd := DeregisterCmd{
				BrokerName: "",
				Context:    nil,
			}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("broker name is required"))
		})
	})
	Describe("Deregister", func() {
		It("Calls the pkg/svcat libs Deregister method with the passed in variables and prints output to the user", func() {
			brokerName := "foobarbroker"
			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.DeregisterReturns(nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := DeregisterCmd{
				Context:    svcattest.NewContext(outputBuffer, fakeApp),
				BrokerName: brokerName,
			}
			err := cmd.Deregister()

			Expect(err).NotTo(HaveOccurred())
			returnedName := fakeSDK.DeregisterArgsForCall(0)
			Expect(returnedName).To(Equal(brokerName))

			output := outputBuffer.String()
			Expect(output).To(Equal("Successfully removed broker \"foobarbroker\""))
		})
	})
})
