package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mssqlServerServiceName             = "csb-azure-mssql-server"
	mssqlServerServiceID               = "a0ab0f36-f8e1-4045-8ddb-1918d2ceafe4"
	mssqlServerServiceDisplayName      = "Deprecated - Azure SQL Server"
	mssqlServerServiceDescription      = "Deprecated - Azure SQL Server (no database attached)"
	mssqlServerServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/sql-database/"
	mssqlServerServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/sql-database/"
)

var _ = Describe("MSSQL Server", Label("MSSQL Server"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mssqlServerServiceName)
		Expect(service.ID).To(Equal(mssqlServerServiceID))
		Expect(service.Description).To(Equal(mssqlServerServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "preview", "deprecated"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(mssqlServerServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(mssqlServerServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(mssqlServerServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("standard"),
					ID:   Equal("1aab10e2-ca79-4755-855a-6073a739d2e0"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(mssqlServerServiceName, "standard", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(mssqlServerServiceName, "standard", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, mssqlServerServiceName, "standard", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})
	})
})
