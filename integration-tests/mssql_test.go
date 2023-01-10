package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mssqlServiceName             = "csb-azure-mssql"
	mssqlServiceID               = "2cfcad84-5824-11ea-b0e2-00155d4dfe6c"
	mssqlServiceDisplayName      = "Azure SQL Database - Single Instance"
	mssqlServiceDescription      = "Azure SQL Database is a fully managed service for the Azure Platform"
	mssqlServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/sql-database/"
	mssqlServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/sql-database/"
)

var _ = Describe("MSSQL", Label("MSSQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mssqlServiceName)
		Expect(service.ID).To(Equal(mssqlServiceID))
		Expect(service.Description).To(Equal(mssqlServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "mssql", "sqlserver", "preview"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(mssqlServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(mssqlServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(mssqlServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small-v2"),
					ID:   Equal("99ed044a-bf9b-11eb-a49a-e347783607d6"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("9295b05a-58c9-11ea-b9df-00155d2c938f"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("9dc5e814-58c9-11ea-9e77-00155d2c938f"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("extra-large"),
					ID:   Equal("a94f7192-5cba-11ea-8b5a-00155d7cdd25"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(mssqlServiceName, "small-v2", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(mssqlServiceName, "small-v2", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, mssqlServiceName, "small-v2", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})
	})
})
