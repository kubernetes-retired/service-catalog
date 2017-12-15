package plugin_client_test

import (
	"errors"

	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client"
	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker", func() {
	var (
		client                *plugin_client.PluginClient
		err                   error
		FakeScClient          *fakes.FakeInterface
		ServicecatalogV1beta1 *fakes.FakeServicecatalogV1beta1Interface
		ClusterServiceBrokers *fakes.FakeClusterServiceBrokerInterface
	)

	BeforeEach(func() {
		client, err = plugin_client.NewClient()
		Expect(err).NotTo(HaveOccurred())
		FakeScClient = &fakes.FakeInterface{}
		ServicecatalogV1beta1 = &fakes.FakeServicecatalogV1beta1Interface{}
		ClusterServiceBrokers = &fakes.FakeClusterServiceBrokerInterface{}

		client.ScClient = FakeScClient
		FakeScClient.ServicecatalogV1beta1Returns(ServicecatalogV1beta1)
		ServicecatalogV1beta1.ClusterServiceBrokersReturns(ClusterServiceBrokers)
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in broker", func() {
			brokerName := "foobar"
			_, err = client.GetBroker(brokerName)

			Expect(ClusterServiceBrokers.GetCallCount()).To(Equal(1))
			name, _ := ClusterServiceBrokers.GetArgsForCall(0)
			Expect(name).To(Equal(brokerName))
		})
		It("Bubbles up errors", func() {
			errorMessage := "broker not found"
			ClusterServiceBrokers.GetReturns(nil, errors.New(errorMessage))

			_, err := client.GetBroker("banana")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServiceBrokers.GetCallCount()).To(Equal(1))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method", func() {
			_, err := client.ListBrokers()

			Expect(err).NotTo(HaveOccurred())
			Expect(ClusterServiceBrokers.ListCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			errorMessage := "foobar"
			ClusterServiceBrokers.ListReturns(nil, errors.New(errorMessage))

			_, err := client.ListBrokers()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServiceBrokers.ListCallCount()).To(Equal(1))
		})
	})
})
