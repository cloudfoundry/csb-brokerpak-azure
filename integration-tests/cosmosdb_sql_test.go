package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	cosmosDBSQLServiceName             = "csb-azure-cosmosdb-sql"
	cosmosDBSQLServiceID               = "685e151f-3ad8-414f-ab5b-54abbb3dee02"
	cosmosDBSQLServiceDisplayName      = "Azure CosmosDB Account - SQL API"
	cosmosDBSQLServiceDescription      = "Azure CosmosDB Account - SQL API"
	cosmosDBSQLServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/cosmos-db/"
	cosmosDBSQLServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/cosmos-db/faq"
)

var _ = Describe("CosmosDB-SQL", Label("CosmosDB-SQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish Azure CosmosDB Account - SQL API in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, cosmosDBSQLServiceName)
		Expect(service.ID).To(Equal(cosmosDBSQLServiceID))
		Expect(service.Description).To(Equal(cosmosDBSQLServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "cosmos", "cosmosdb", "cosmos-sql", "cosmosdb-sql", "preview"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(cosmosDBSQLServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(cosmosDBSQLServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(cosmosDBSQLServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small"),
					ID:   Equal("ca38881c-2d6b-4db4-988c-d8e49f3293da"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("f666cd68-cbba-4c58-b532-bfc0cb533011"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("2d5ee55d-1315-40ca-a9e9-08f4f76e880f"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		planName := "large"
		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(cosmosDBSQLServiceName, planName, params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"invalid region",
				map[string]any{"location": "-Asia-northeast1"},
				"location: Does not match pattern '^[a-z][a-z0-9]+$'",
			),
			Entry(
				"instance name maximum length is 44 characters",
				map[string]any{"instance_name": stringOfLen(45)},
				"instance_name: String length must be less than or equal to 44",
			),
			Entry(
				"resource group name maximum length is 64 characters",
				map[string]any{"resource_group": stringOfLen(65)},
				"resource_group: String length must be less than or equal to 64",
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(cosmosDBSQLServiceName, planName, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("request_units", BeNumerically("==", 10000)),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb%s", instanceID)),
					HaveKeyWithValue("failover_locations", ConsistOf("westus", "eastus")),
					HaveKeyWithValue("enable_multiple_write_locations", BeTrue()),
					HaveKeyWithValue("enable_automatic_failover", BeTrue()),
					HaveKeyWithValue("resource_group", BeEmpty()),
					HaveKeyWithValue("db_name", fmt.Sprintf("csb-db%s", instanceID)),
					HaveKeyWithValue("location", "westus"),
					HaveKeyWithValue("ip_range_filter", "0.0.0.0"),
					HaveKeyWithValue("consistency_level", "Session"),
					HaveKeyWithValue("max_interval_in_seconds", BeNumerically("==", 5)),
					HaveKeyWithValue("max_staleness_prefix", BeNumerically("==", 100)),
					HaveKeyWithValue("skip_provider_registration", BeFalse()),
					HaveKeyWithValue("authorized_network", BeEmpty()),
					HaveKeyWithValue("private_endpoint_subnet_id", BeEmpty()),
					HaveKeyWithValue("private_dns_zone_ids", BeEmpty()),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(cosmosDBSQLServiceName, planName, map[string]any{
				"instance_name":              "my-cosmosdb-sql",
				"resource_group":             "my-resource-group",
				"db_name":                    "my-db-name",
				"location":                   "uksouth",
				"private_endpoint_subnet_id": "subnet-id",
				"private_dns_zone_ids":       []string{"dns-zone-id-1", "dns-zone-id-2"},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("instance_name", "my-cosmosdb-sql"),
					HaveKeyWithValue("resource_group", "my-resource-group"),
					HaveKeyWithValue("db_name", "my-db-name"),
					HaveKeyWithValue("location", "uksouth"),
					HaveKeyWithValue("private_endpoint_subnet_id", "subnet-id"),
					HaveKeyWithValue("private_dns_zone_ids", ConsistOf("dns-zone-id-1", "dns-zone-id-2")),
				),
			)
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(cosmosDBSQLServiceName, "small", nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, cosmosDBSQLServiceName, "small", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			const initialProvisionInvocation = 1
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
		})
	})
})
