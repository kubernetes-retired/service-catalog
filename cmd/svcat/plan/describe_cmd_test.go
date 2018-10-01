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

package plan

import (
	"github.com/kubernetes-incubator/service-catalog/cmd/svcat/command"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Describe Command", func() {
	Describe("NewDescribeCmd", func() {
		It("Builds and returns a cobra command", func() {
			cxt := &command.Context{}
			cmd := NewDescribeCmd(cxt)
			Expect(*cmd).NotTo(BeNil())
			Expect(cmd.Use).To(Equal("plan [NAME]"))
			Expect(cmd.Short).To(ContainSubstring("Show details of a specific plan"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan standard800"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan --uuid 08e4b43a-36bc-447e-a81f-8202b13e339c"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan PLAN_NAME --scope cluster"))
			Expect(cmd.Example).To(ContainSubstring("svcat describe plan PLAN_NAME --scope namespace --namespace NAMESPACE_NAME"))
			Expect(len(cmd.Aliases)).To(Equal(2))

			uuidFlag := cmd.Flags().Lookup("uuid")
			Expect(uuidFlag).NotTo(BeNil())
			Expect(uuidFlag.Usage).To(ContainSubstring("Whether or not to get the class by UUID (the default is by name)"))

			showSchemaFlag := cmd.Flags().Lookup("show-schemas")
			Expect(showSchemaFlag).NotTo(BeNil())
			Expect(showSchemaFlag.Usage).To(ContainSubstring("Whether or not to show instance and binding parameter schemas"))

			scopeFlag := cmd.Flags().Lookup("scope")
			Expect(scopeFlag).NotTo(BeNil())
			Expect(scopeFlag.Usage).To(ContainSubstring("Limit the results to a particular scope"))

			namespaceFlag := cmd.Flags().Lookup("namespace")
			Expect(namespaceFlag).NotTo(BeNil())
			Expect(namespaceFlag.Usage).To(ContainSubstring("If present, the namespace scope for this request"))
		})
	})
	Describe("Validate", func() {
		It("errors if no argument is provided", func() {
			cmd := DescribeCmd{}
			err := cmd.Validate([]string{})
			Expect(err).To(HaveOccurred())
		})
	})
	/*
		Describe("Run", func() {
			It("Calls the pkg/svcat libs CreateClassFrom method with the passed in variables for a cluster class and prints output to the user", func() {
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
						Name:      className,
						Namespace: classNamespace,
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
	*/
})
