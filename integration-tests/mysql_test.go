package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mysqlServiceName             = "csb-azure-mysql"
	mysqlServiceID               = "cac4a46b-c4ec-49df-9b11-06457a29d31e"
	mysqlServiceDisplayName      = "Retired - Azure Database for MySQL single servers"
	mysqlServiceDescription      = "Retired - Azure Database for MySQL single servers"
	mysqlServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/mysql/"
	mysqlServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/mysql/"
)

var _ = Describe("MySQL", Label("MySQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mysqlServiceName)
		Expect(service.ID).To(Equal(mysqlServiceID))
		Expect(service.Description).To(Equal(mysqlServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "mysql", "preview", "retired", "single server"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(mysqlServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(mysqlServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(mysqlServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small"),
					ID:   Equal("828e324e-6b34-4f50-b224-9b956dd2d1b7"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("9eb836dd-4B90-4cF7-bc06-1986103802d3"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("6f8Ea44c-6840-4b0b-9068-f0cd9b17437c"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(mysqlServiceName, "small", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(mysqlServiceName, "small", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, mysqlServiceName, "small", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})
	})
})
