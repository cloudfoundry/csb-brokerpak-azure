package terraformtests

import (
	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/gomega/gstruct"

	. "csbbrokerpakazure/terraform-tests/helpers"

	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL Flexible Server", Label("postgresql-flexible-server"), Ordered, func() {
	const (
		instanceName      = "csb-postgresql-flexible-server-instance"
		resourceGroupName = "csb-resource-group"
		dbName            = "csb-db"
		subnetID          = "/subscriptions/azureSubscriptionID/resourceGroups/csb-resource-group/providers/Microsoft.Network/virtualNetworks/csb-resource-group-vnet/subnets/subnet-name"
		privateDNSZoneID  = "/subscriptions/azureSubscriptionID/resourceGroups/csb-resource-group/providers/Microsoft.Network/privateDnsZones/test.postgres.database.azure.com"
	)

	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
		defaultVars           map[string]any
	)

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "azure-postgresql-flexible-server/provision")
		Init(terraformProvisionDir)

		defaultVars = map[string]any{
			"azure_client_id":                  azureClientID,
			"azure_client_secret":              azureClientSecret,
			"azure_subscription_id":            azureSubscriptionID,
			"azure_tenant_id":                  azureTenantID,
			"instance_name":                    instanceName,
			"db_name":                          dbName,
			"location":                         "westus",
			"labels":                           map[string]any{"k1": "v1"},
			"storage_gb":                       32,
			"resource_group":                   resourceGroupName,
			"postgres_version":                 "11",
			"sku_name":                         "GP_Standard_D2ads_v5",
			"allow_access_from_azure_services": true,
			"delegated_subnet_id":              nil,
			"private_dns_zone_id":              nil,
			"private_endpoint_subnet_id":       "",
		}
	})

	Context("with Default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(5))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_postgresql_flexible_server",
				"azurerm_postgresql_flexible_server_database",
				"azurerm_postgresql_flexible_server_firewall_rule",
				"random_string",
				"random_password",
			))
		})

		It("should create a postgresql flexible server with the right values", func() {
			Expect(AfterValuesForType(plan, "azurerm_postgresql_flexible_server")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                Equal(instanceName),
					"resource_group_name": Equal(resourceGroupName),
					"location":            Equal("westus"),
					"version":             Equal("11"),
					"sku_name":            Equal("GP_Standard_D2ads_v5"),
					"storage_mb":          BeNumerically("==", 32768),
					"delegated_subnet_id": BeNil(),

					"tags": MatchAllKeys(Keys{
						"k1": Equal("v1"),
					}),
				}),
			)
		})

		It("should create a postgresql flexible database with the right values", func() {
			Expect(AfterValuesForType(plan, "azurerm_postgresql_flexible_server_database")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":      Equal(dbName),
					"charset":   Equal("UTF8"),
					"collation": Equal("en_US.utf8"),
				}))
		})
	})

	When("no resource group is passed", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"resource_group": "",
			}))
		})

		It("should create a resource group", func() {
			Expect(plan.ResourceChanges).To(HaveLen(6))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_resource_group",
				"azurerm_postgresql_flexible_server",
				"azurerm_postgresql_flexible_server_database",
				"azurerm_postgresql_flexible_server_firewall_rule",
				"random_string",
				"random_password",
			))

			Expect(AfterValuesForType(plan, "azurerm_resource_group")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name": Equal("rg-csb-postgresql-flexible-server-instance"),
				}))
		})
	})

	When("private access with virtual network is enabled", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"delegated_subnet_id":              subnetID,
				"private_dns_zone_id":              privateDNSZoneID,
				"allow_access_from_azure_services": false,
			}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(4))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_postgresql_flexible_server",
				"azurerm_postgresql_flexible_server_database",
				"random_string",
				"random_password",
			))
		})

		It("should setup network delegation", func() {
			Expect(AfterValuesForType(plan, "azurerm_postgresql_flexible_server")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                Equal(instanceName),
					"resource_group_name": Equal(resourceGroupName),
					"location":            Equal("westus"),
					"version":             Equal("11"),
					"sku_name":            Equal("GP_Standard_D2ads_v5"),
					"storage_mb":          BeNumerically("==", 32768),
					"delegated_subnet_id": Equal(subnetID),
					"private_dns_zone_id": Equal(privateDNSZoneID),

					"tags": MatchAllKeys(Keys{
						"k1": Equal("v1"),
					}),
				}),
			)
		})
	})

	When("private endpoint is enabled", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"private_endpoint_subnet_id":       subnetID,
				"private_dns_zone_id":              privateDNSZoneID,
				"allow_access_from_azure_services": false,
			}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(5))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_postgresql_flexible_server",
				"azurerm_postgresql_flexible_server_database",
				"random_string",
				"random_password",
				"azurerm_private_endpoint",
			))
		})

		It("should setup a new private endpoint", func() {
			Expect(AfterValuesForType(plan, "azurerm_private_endpoint")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                Equal("csb-postgresql-flexible-server-instance-private_endpoint"),
					"location":            Equal("westus"),
					"resource_group_name": Equal(resourceGroupName),
					"subnet_id":           Equal(subnetID),
					"tags": MatchAllKeys(Keys{
						"k1": Equal("v1"),
					}),
					"private_service_connection": ConsistOf(MatchKeys(IgnoreExtras, Keys{
						"name":                 Equal("csb-postgresql-flexible-server-instance-private_service_connection"),
						"is_manual_connection": BeFalse(),
						"subresource_names":    ConsistOf("postgresqlServer"),
					})),
					"private_dns_zone_group": ConsistOf(MatchKeys(IgnoreExtras, Keys{
						"private_dns_zone_ids": ConsistOf(privateDNSZoneID),
					})),
				}),
			)
		})
	})
})
