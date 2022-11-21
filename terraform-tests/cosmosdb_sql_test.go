package terraformtests

import (
	"path"

	. "csbbrokerpakazure/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("CosmosDB SQL", Label("cosmosdb-sql-terraform"), Ordered, func() {
	const (
		instanceName      = "csb-cosmosdb-sql"
		resourceGroupName = "csb-resource-group"
		dbName            = "csb-db"
	)

	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"azure_client_id":                 azureClientID,
		"azure_client_secret":             azureClientSecret,
		"azure_subscription_id":           azureSubscriptionID,
		"azure_tenant_id":                 azureTenantID,
		"request_units":                   10000,
		"instance_name":                   instanceName,
		"failover_locations":              []string{"westus", "eastus"},
		"enable_multiple_write_locations": true,
		"enable_automatic_failover":       true,
		"resource_group":                  resourceGroupName,
		"db_name":                         dbName,
		"location":                        "westus",
		"ip_range_filter":                 "0.0.0.0",
		"consistency_level":               "BoundedStaleness",
		"max_interval_in_seconds":         5,
		"max_staleness_prefix":            100,
		"skip_provider_registration":      false,
		"authorized_network":              "",
		"labels":                          map[string]any{"k1": "v1"},
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "azure-cosmosdb")
		Init(terraformProvisionDir)
	})

	Context("with Default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(2))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_cosmosdb_account",
				"azurerm_cosmosdb_sql_database",
			))
		})

		It("should create a cosmosdb account with the right values", func() {
			Expect(AfterValuesForType(plan, "azurerm_cosmosdb_account")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                              Equal(instanceName),
					"location":                          Equal("westus"),
					"resource_group_name":               Equal(resourceGroupName),
					"offer_type":                        Equal("Standard"),
					"kind":                              Equal("GlobalDocumentDB"),
					"enable_automatic_failover":         BeTrue(),
					"enable_multiple_write_locations":   BeTrue(),
					"is_virtual_network_filter_enabled": BeFalse(),
					"ip_range_filter":                   Equal("0.0.0.0"),
					"tags": MatchAllKeys(Keys{
						"k1": Equal("v1"),
					}),

					"consistency_policy": ConsistOf(
						MatchKeys(IgnoreExtras, Keys{
							"consistency_level":       Equal("BoundedStaleness"),
							"max_interval_in_seconds": BeNumerically("==", 5),
							"max_staleness_prefix":    BeNumerically("==", 100),
						}),
					),

					"geo_location": ConsistOf(
						MatchKeys(IgnoreExtras, Keys{
							"location":          Equal("westus"),
							"failover_priority": BeNumerically("==", 0),
						}),
						MatchKeys(IgnoreExtras, Keys{
							"location":          Equal("eastus"),
							"failover_priority": BeNumerically("==", 1),
						}),
					),
				}),
			)
		})

		It("should create a cosmosdb sql database with the right values", func() {
			Expect(AfterValuesForType(plan, "azurerm_cosmosdb_sql_database")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                Equal(dbName),
					"resource_group_name": Equal(resourceGroupName),
					"account_name":        Equal(instanceName),
					"throughput":          BeNumerically("==", 10000),
				}))
		})
	})
})
