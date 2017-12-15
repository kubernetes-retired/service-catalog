package plugin_client_test

import (
	"errors"

	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client"
	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Binding", func() {
	var (
		client                *plugin_client.PluginClient
		err                   error
		FakeScClient          *fakes.FakeInterface
		ServicecatalogV1beta1 *fakes.FakeServicecatalogV1beta1Interface
		ServiceBindings       *fakes.FakeServiceBindingInterface
	)

	BeforeEach(func() {
		client, err = plugin_client.NewClient()
		Expect(err).NotTo(HaveOccurred())
		FakeScClient = &fakes.FakeInterface{}
		ServicecatalogV1beta1 = &fakes.FakeServicecatalogV1beta1Interface{}
		ServiceBindings = &fakes.FakeServiceBindingInterface{}

		client.ScClient = FakeScClient
		FakeScClient.ServicecatalogV1beta1Returns(ServicecatalogV1beta1)
		ServicecatalogV1beta1.ServiceBindingsReturns(ServiceBindings)
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in binding and namespace", func() {
			namespace := "foobar_namespace"
			bindingName := "potato_binding"

			_, err := client.GetBinding(bindingName, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(ServicecatalogV1beta1.ServiceBindingsArgsForCall(0)).To(Equal(namespace))
			returnedName, _ := ServiceBindings.GetArgsForCall(0)
			Expect(returnedName).To(Equal(bindingName))
			Expect(ServiceBindings.GetCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			namespace := "foobar_namespace"
			bindingName := "potato_binding"
			errorMessage := "binding not found"

			ServiceBindings.GetReturns(nil, errors.New(errorMessage))
			_, err := client.GetBinding(bindingName, namespace)
			Expect(err).To(HaveOccurred())
			Expect(ServiceBindings.GetCallCount()).To(Equal(1))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method with the specified namespace", func() {
			namespace := "foobar_namespace"

			_, err := client.ListBindings(namespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(ServicecatalogV1beta1.ServiceBindingsArgsForCall(0)).To(Equal(namespace))
			Expect(ServiceBindings.ListCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			ServiceBindings.ListReturns(nil, errors.New("foobar"))
			namespace := "foobar_namespace"

			_, err := client.ListBindings(namespace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("foobar"))
			Expect(ServicecatalogV1beta1.ServiceBindingsArgsForCall(0)).To(Equal(namespace))
			Expect(ServiceBindings.ListCallCount()).To(Equal(1))
		})
	})
})
