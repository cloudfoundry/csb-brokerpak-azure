package terraformtests

import (
	"path"

	. "csbbrokerpakazure/terraform-tests/helpers"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Redis", Label("redis"), Ordered, func() {
	const (
		instanceName      = "csb-redis"
		resourceGroupName = "csb-redis-rg"
	)

	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"azure_client_id":            azureClientID,
		"azure_client_secret":        azureClientSecret,
		"azure_subscription_id":      azureSubscriptionID,
		"azure_tenant_id":            azureTenantID,
		"instance_name":              instanceName,
		"resource_group":             resourceGroupName,
		"location":                   "westus",
		"family":                     "C",
		"sku_name":                   "Basic",
		"capacity":                   1,
		"redis_version":              "4",
		"firewall_rules":             [][]string{},
		"maxmemory_policy":           "",
		"private_dns_zone_ids":       []string{},
		"private_endpoint_subnet_id": "",
		"subnet_id":                  "",
		"tls_min_version":            "",
		"skip_provider_registration": false,

		"labels": map[string]any{"k1": "v1"},
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "azure-redis/provision")
		Init(terraformProvisionDir)
	})

	Context("with Default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(2))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_redis_cache",
				"random_string",
			))
		})

		It("should create a redis cache with the right values", func() {
			Expect(AfterValuesForType(plan, "azurerm_redis_cache")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name":                          Equal(instanceName),
					"sku_name":                      Equal("Basic"),
					"redis_version":                 Equal("4"),
					"family":                        Equal("C"),
					"capacity":                      BeNumerically("==", 1),
					"location":                      Equal("westus"),
					"resource_group_name":           Equal(resourceGroupName),
					"minimum_tls_version":           Equal("1.2"),
					"public_network_access_enabled": BeTrue(),
					"tags": MatchAllKeys(Keys{
						"k1": Equal("v1"),
					}),
					"redis_configuration": ConsistOf(MatchKeys(IgnoreExtras, Keys{
						"maxmemory_policy": Equal("allkeys-lru"),
					})),
				}),
			)
		})
	})

	When("no resource group is passed", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"resource_group": "",
			}))
		})

		It("should create a resource group", func() {
			Expect(plan.ResourceChanges).To(HaveLen(3))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_resource_group",
				"azurerm_redis_cache",
				"random_string",
			))

			Expect(AfterValuesForType(plan, "azurerm_resource_group")).To(
				MatchKeys(IgnoreExtras, Keys{
					"name": Equal("rg-csb-redis"),
				}))
		})
	})

	When("firewall rules are set", func() {
		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"firewall_rules": [][]string{{"1.2.3.4", "2.3.4.5"}, {"5.6.7.8", "6.7.8.9"}},
			}))
		})

		It("should create a firewall rule for each", func() {
			Expect(plan.ResourceChanges).To(HaveLen(4))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_redis_cache",
				"azurerm_redis_firewall_rule",
				"azurerm_redis_firewall_rule",
				"random_string",
			))
		})
	})

	When("private endpoint is enabled", func() {
		var subnetID = "/subscriptions/azureSubscriptionID/resourceGroups/csb-redis-rg/providers/Microsoft.Network/virtualNetworks/csb-redis-rg-platform/subnets/csb-redis-rg-pas-subnet"
		var dnsID = "/subscriptions/azureSubscriptionID/resourceGroups/dns-configuration/providers/Microsoft.Network/virtualNetworks/dnszones/test"

		BeforeEach(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"private_endpoint_subnet_id": subnetID,
				"private_dns_zone_ids":       []string{dnsID},
			}))
		})

		It("should create a private endpoint", func() {
			Expect(plan.ResourceChanges).To(HaveLen(3))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"azurerm_redis_cache",
				"azurerm_private_endpoint",
				"random_string",
			))

			Expect(AfterValuesForType(plan, "azurerm_private_endpoint")).To(
				MatchKeys(IgnoreExtras, Keys{
					"location":            Equal("westus"),
					"resource_group_name": Equal(resourceGroupName),
					"subnet_id":           Equal(subnetID),
					"tags": MatchAllKeys(Keys{
						"k1": Equal("v1"),
					}),
					"private_service_connection": ConsistOf(MatchKeys(IgnoreExtras, Keys{
						"is_manual_connection": BeFalse(),
						"subresource_names":    ConsistOf("redisCache"),
					})),
					"private_dns_zone_group": ConsistOf(MatchKeys(IgnoreExtras, Keys{
						"private_dns_zone_ids": ConsistOf(dnsID),
					})),
				}),
			)
		})
	})
})
