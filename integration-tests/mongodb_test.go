package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mongoDBServiceName             = "csb-azure-mongodb"
	mongoDBServiceID               = "e5d2898e-534a-11ea-a4e8-00155da9789e"
	mongoDBServiceDisplayName      = "Azure Cosmos DB's API for MongoDB"
	mongoDBServiceDescription      = "The Cosmos DB service implements wire protocols for MongoDB.  Azure Cosmos DB is Microsoft's globally distributed, multi-model database service for mission-critical application"
	mongoDBServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/cosmos-db/mongodb-introduction"
	mongoDBServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/cosmos-db/faq"
)

var _ = Describe("MongoDB", Label("MongoDB"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mongoDBServiceName)
		Expect(service.ID).To(Equal(mongoDBServiceID))
		Expect(service.Description).To(Equal(mongoDBServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "mongodb", "preview", "cosmosdb-mongo", "cosmosdb-mongodb"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(mongoDBServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(mongoDBServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(mongoDBServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small"),
					ID:   Equal("4ba45322-534c-11ea-b346-00155da9789e"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("6a28ad34-534c-11ea-9bac-00155da9789e"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("72227eac-534c-11ea-b7ca-00155da9789e"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(mongoDBServiceName, "small", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("provisioning with custom shard_key", func() {
		It("should affect the value of unique_indexes computed value", func() {
			_, err := broker.Provision(mongoDBServiceName, "small", map[string]any{"shard_key": "shard"})
			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("shard_key", "shard"),
					HaveKeyWithValue("unique_indexes", "_id,shard"),
				),
			)
		})
	})

	Describe("provisioning with custom unique_indexes", func() {
		It("should override computed unique_indexes", func() {
			_, err := broker.Provision(mongoDBServiceName, "small", map[string]any{"unique_indexes": "uidx1,uidx2", "shard_key": "shard"})
			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("unique_indexes", "uidx1,uidx2"),
				),
			)
		})
	})

	Describe("provisioning with default values", func() {
		It("should use default values for shard_key, indexes and unique_indexes", func() {
			_, err := broker.Provision(mongoDBServiceName, "small", map[string]any{})
			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("indexes", ""),
					HaveKeyWithValue("shard_key", "uniqueKey"),
					HaveKeyWithValue("unique_indexes", "_id,uniqueKey"),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(mongoDBServiceName, "small", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, mongoDBServiceName, "small", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})
	})
})
