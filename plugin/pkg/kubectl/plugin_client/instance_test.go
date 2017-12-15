package plugin_client_test

import (
	"errors"

	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client"
	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instances", func() {
	var (
		client                *plugin_client.PluginClient
		err                   error
		FakeScClient          *fakes.FakeInterface
		ServicecatalogV1beta1 *fakes.FakeServicecatalogV1beta1Interface
		ServiceInstances      *fakes.FakeServiceInstanceInterface
	)

	BeforeEach(func() {
		client, err = plugin_client.NewClient()
		Expect(err).NotTo(HaveOccurred())
		FakeScClient = &fakes.FakeInterface{}
		ServicecatalogV1beta1 = &fakes.FakeServicecatalogV1beta1Interface{}
		ServiceInstances = &fakes.FakeServiceInstanceInterface{}

		client.ScClient = FakeScClient
		FakeScClient.ServicecatalogV1beta1Returns(ServicecatalogV1beta1)
		ServicecatalogV1beta1.ServiceInstancesReturns(ServiceInstances)
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in instance and namespace", func() {
			namespace := "foobar_namespace"
			instanceName := "potato_instance"

			_, err := client.GetInstance(instanceName, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(ServicecatalogV1beta1.ServiceInstancesArgsForCall(0)).To(Equal(namespace))
			returnedName, _ := ServiceInstances.GetArgsForCall(0)
			Expect(returnedName).To(Equal(instanceName))
			Expect(ServiceInstances.GetCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			namespace := "foobar_namespace"
			instanceName := "potato_instance"
			errorMessage := "instance not found"

			ServiceInstances.GetReturns(nil, errors.New(errorMessage))
			_, err := client.GetInstance(instanceName, namespace)
			Expect(err).To(HaveOccurred())
			Expect(ServiceInstances.GetCallCount()).To(Equal(1))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method with the specified namespace", func() {
			namespace := "foobar_namespace"

			_, err := client.ListInstances(namespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(ServicecatalogV1beta1.ServiceInstancesArgsForCall(0)).To(Equal(namespace))
			Expect(ServiceInstances.ListCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			ServiceInstances.ListReturns(nil, errors.New("foobar"))
			namespace := "foobar_namespace"

			_, err := client.ListInstances(namespace)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("foobar"))
			Expect(ServicecatalogV1beta1.ServiceInstancesArgsForCall(0)).To(Equal(namespace))
			Expect(ServiceInstances.ListCallCount()).To(Equal(1))
		})
	})
})
