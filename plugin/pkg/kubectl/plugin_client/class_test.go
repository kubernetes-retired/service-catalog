package plugin_client_test

import (
	"errors"

	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client"
	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Class", func() {
	var (
		client                *plugin_client.PluginClient
		err                   error
		FakeScClient          *fakes.FakeInterface
		ServicecatalogV1beta1 *fakes.FakeServicecatalogV1beta1Interface
		ClusterServiceClasses *fakes.FakeClusterServiceClassInterface
	)

	BeforeEach(func() {
		client, err = plugin_client.NewClient()
		Expect(err).NotTo(HaveOccurred())
		FakeScClient = &fakes.FakeInterface{}
		ServicecatalogV1beta1 = &fakes.FakeServicecatalogV1beta1Interface{}
		ClusterServiceClasses = &fakes.FakeClusterServiceClassInterface{}

		client.ScClient = FakeScClient
		FakeScClient.ServicecatalogV1beta1Returns(ServicecatalogV1beta1)
		ServicecatalogV1beta1.ClusterServiceClassesReturns(ClusterServiceClasses)
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in class", func() {
			className := "foobar"
			_, err = client.GetClass(className)

			Expect(ClusterServiceClasses.GetCallCount()).To(Equal(1))
			name, _ := ClusterServiceClasses.GetArgsForCall(0)
			Expect(name).To(Equal(className))
		})
		It("Bubbles up errors", func() {
			errorMessage := "class not found"
			ClusterServiceClasses.GetReturns(nil, errors.New(errorMessage))

			_, err := client.GetClass("banana")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServiceClasses.GetCallCount()).To(Equal(1))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method", func() {
			_, err := client.ListClasses()

			Expect(err).NotTo(HaveOccurred())
			Expect(ClusterServiceClasses.ListCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			errorMessage := "foobar"
			ClusterServiceClasses.ListReturns(nil, errors.New(errorMessage))

			_, err := client.ListClasses()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServiceClasses.ListCallCount()).To(Equal(1))
		})
	})
})
