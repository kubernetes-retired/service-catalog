package plugin_client_test

import (
	"errors"

	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client"
	"github.com/kubernetes-incubator/service-catalog/plugin/pkg/kubectl/plugin_client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plan", func() {
	var (
		client                *plugin_client.PluginClient
		err                   error
		FakeScClient          *fakes.FakeInterface
		ServicecatalogV1beta1 *fakes.FakeServicecatalogV1beta1Interface
		ClusterServicePlans   *fakes.FakeClusterServicePlanInterface
	)

	BeforeEach(func() {
		client, err = plugin_client.NewClient()
		Expect(err).NotTo(HaveOccurred())
		FakeScClient = &fakes.FakeInterface{}
		ServicecatalogV1beta1 = &fakes.FakeServicecatalogV1beta1Interface{}
		ClusterServicePlans = &fakes.FakeClusterServicePlanInterface{}

		client.ScClient = FakeScClient
		FakeScClient.ServicecatalogV1beta1Returns(ServicecatalogV1beta1)
		ServicecatalogV1beta1.ClusterServicePlansReturns(ClusterServicePlans)
	})

	Describe("Get", func() {
		It("Calls the generated v1beta1 List method with the passed in plan", func() {
			planName := "foobar"
			_, err = client.GetPlan(planName)

			Expect(ClusterServicePlans.GetCallCount()).To(Equal(1))
			name, _ := ClusterServicePlans.GetArgsForCall(0)
			Expect(name).To(Equal(planName))
		})
		It("Bubbles up errors", func() {
			errorMessage := "plan not found"
			ClusterServicePlans.GetReturns(nil, errors.New(errorMessage))

			_, err := client.GetPlan("banana")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServicePlans.GetCallCount()).To(Equal(1))
		})
	})

	Describe("List", func() {
		It("Calls the generated v1beta1 List method", func() {
			_, err := client.ListPlans()

			Expect(err).NotTo(HaveOccurred())
			Expect(ClusterServicePlans.ListCallCount()).To(Equal(1))
		})
		It("Bubbles up errors", func() {
			errorMessage := "foobar"
			ClusterServicePlans.ListReturns(nil, errors.New(errorMessage))

			_, err := client.ListPlans()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(errorMessage))
			Expect(ClusterServicePlans.ListCallCount()).To(Equal(1))
		})
	})
})
