package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mssqlFailoverGroupServiceName             = "csb-azure-mssql-failover-group"
	mssqlFailoverGroupServiceID               = "76d0e602-2b79-4c1e-bbbe-03913a1cfda2"
	mssqlFailoverGroupServiceDisplayName      = "Azure SQL Failover Group"
	mssqlFailoverGroupServiceDescription      = "Manages auto failover group for managed Azure SQL on the Azure Platform"
	mssqlFailoverGroupServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/sql-database/sql-database-auto-failover-group/"
	mssqlFailoverGroupServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/sql-database/sql-database-auto-failover-group/"
)

var _ = Describe("MSSQL Auto-failover group", Label("MSSQL Auto-failover group"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mssqlFailoverGroupServiceName)
		Expect(service.ID).To(Equal(mssqlFailoverGroupServiceID))
		Expect(service.Description).To(Equal(mssqlFailoverGroupServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "mssql", "sqlserver", "dr", "failover", "preview"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(mssqlFailoverGroupServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(mssqlFailoverGroupServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(mssqlFailoverGroupServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small-v2"),
					ID:   Equal("eb9856fa-b285-11eb-ae46-536679aeffe8"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("c5e8ec57-ab5a-4bbf-ac6d-5075a97ed1a5"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("605a7a26-b1dd-4ce5-a382-4233e98469a8"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(mssqlFailoverGroupServiceName, "small-v2", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(mssqlFailoverGroupServiceName, "small-v2", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, mssqlFailoverGroupServiceName, "small-v2", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})
	})
})
