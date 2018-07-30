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

package class_test

import (
	"bytes"

	. "github.com/kubernetes-incubator/service-catalog/cmd/svcat/class"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	servicecatalogfakes "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog/service-catalogfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Create Command", func() {
	Describe("NewCreateCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewCreateCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("class [NAME] --from [EXISTING_NAME]"))
			Expect(cmd.Short).To(ContainSubstring("Copies an existing class into a new user-defined cluster-scoped class"))
			Expect(cmd.Example).To(ContainSubstring("svcat create class newclass --from mysqldb"))
			Expect(len(cmd.Aliases)).To(Equal(0))
		})
	})
	Describe("Validate name is provided", func() {
		It("errors if a new class name is not provided", func() {
			cmd := CreateCmd{
				Context: nil,
				Name:    "",
				From:    "existingclass",
			}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Validate from is provided", func() {
		It("errors if a existing class name is not provided using from", func() {
			cmd := CreateCmd{
				Context: nil,
				Name:    "newclass",
				From:    "",
			}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Create", func() {
		It("Calls the pkg/svcat libs Create method with the passed in variables and prints output to the user", func() {
			className := "newclass"
			existingClassName := "existingclass"

			classToReturn := &v1beta1.ClusterServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Name: className,
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.CreateClassReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := CreateCmd{
				Context: svcattest.NewContext(outputBuffer, fakeApp),
				Name:    className,
				From:    existingClassName,
			}
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			class := fakeSDK.CreateClassArgsForCall(0)
			Expect(class.Name).To(Equal(className))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(className))
		})
	})
})
