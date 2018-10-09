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

package class

import (
	"bytes"

	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/test"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/svcat"
	servicecatalog "github.com/kubernetes-incubator/service-catalog/pkg/svcat/service-catalog"
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
			Expect(cmd.Example).To(ContainSubstring("svcat create class newclass --from mysqldb --scope cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat create class newclass --from mysqldb --scope namespace --namespace newnamespace"))
			Expect(len(cmd.Aliases)).To(Equal(0))

			fromFlag := cmd.Flags().Lookup("from")
			Expect(fromFlag).NotTo(BeNil())
			Expect(fromFlag.Usage).To(ContainSubstring("Name from an existing class that will be copied (Required)"))

			scopeFlag := cmd.Flags().Lookup("scope")
			Expect(scopeFlag).NotTo(BeNil())
			Expect(scopeFlag.Usage).To(ContainSubstring("Limit the command to a particular scope"))

			namespaceFlag := cmd.Flags().Lookup("namespace")
			Expect(namespaceFlag).NotTo(BeNil())
			Expect(namespaceFlag.Usage).To(ContainSubstring("If present, the namespace scope for this request"))
		})
	})
	Describe("Validate", func() {
		It("errors if no argument is provided", func() {
			cmd := CreateCmd{
				Name: "",
				From: "class",
			}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Run", func() {
		It("Calls the pkg/svcat libs CreateClassFrom method with the passed in variables for a cluster class and prints output to the user", func() {
			className := "newclass"
			existingClassName := "existingclass"

			classToReturn := &v1beta1.ClusterServiceClass{
				Spec: v1beta1.ClusterServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: className,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.CreateClassFromReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := CreateCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
				Name:       className,
				From:       existingClassName,
			}
			cmd.Scope = servicecatalog.ClusterScope
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			opts := fakeSDK.CreateClassFromArgsForCall(0)
			Expect(opts.Name).To(Equal(className))
			Expect(opts.From).To(Equal(existingClassName))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(className))
		})
		It("Calls the pkg/svcat libs CreateClassFrom method with the passed in variables for a namespace class and prints output to the user", func() {
			className := "newclass"
			classNamespace := "default"
			existingClassName := "existingclass"

			classToReturn := &v1beta1.ServiceClass{
				ObjectMeta: v1.ObjectMeta{
					Namespace: classNamespace,
				},
				Spec: v1beta1.ServiceClassSpec{
					CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
						ExternalName: className,
					},
				},
			}

			outputBuffer := &bytes.Buffer{}

			fakeApp, _ := svcat.NewApp(nil, nil, "default")
			fakeSDK := new(servicecatalogfakes.FakeSvcatClient)
			fakeSDK.CreateClassFromReturns(classToReturn, nil)
			fakeApp.SvcatClient = fakeSDK
			cmd := CreateCmd{
				Namespaced: &command.Namespaced{Context: svcattest.NewContext(outputBuffer, fakeApp)},
				Scoped:     command.NewScoped(),
				Name:       className,
				From:       existingClassName,
			}
			cmd.Scope = servicecatalog.NamespaceScope
			cmd.Namespace = classNamespace
			err := cmd.Run()

			Expect(err).NotTo(HaveOccurred())
			opts := fakeSDK.CreateClassFromArgsForCall(0)
			Expect(opts.Name).To(Equal(className))
			Expect(opts.From).To(Equal(existingClassName))

			output := outputBuffer.String()
			Expect(output).To(ContainSubstring(className))
			Expect(output).To(ContainSubstring(classNamespace))
		})
	})
})
