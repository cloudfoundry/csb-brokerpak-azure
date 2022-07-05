package csbmssqldbrunfailover_test

import (
	"fmt"
	"regexp"

	"csbbrokerpakazure/providers/terraform-provider-csbmssqldbrunfailover/csbmssqldbrunfailover"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Provider Configuration", func() {
	var (
		azureTenantID,
		azureClientID,
		azureClientSecret,
		azureSubscriptionID,
		resourceGroup,
		partnerServerResourceGroup,
		serverName,
		partnerServerName,
		failoverGroup string
	)

	BeforeEach(func() {

		azureTenantID = "some-tenant-id"
		azureClientID = "some-client-id"
		azureClientSecret = "client-secret"
		azureSubscriptionID = "subscription-id"
		resourceGroup = "resource-group"
		partnerServerResourceGroup = "partner-resource-group"
		serverName = "server-name"
		partnerServerName = "partner-server-name"
		failoverGroup = "failover-group"

	})

	DescribeTable(
		"validation of parameters",
		func(cb func(), expectError string) {
			cb()

			hcl := fmt.Sprintf(`
				provider "csbmssqldbrunfailover" {
				  azure_tenant_id       = "%s"
				  azure_client_id       = "%s"
				  azure_client_secret   = "%s"
				  azure_subscription_id = "%s"
				}
				
				resource "csbmssqldbrunfailover_failover" "failover" {
				  resource_group                = "%s"
				  partner_server_resource_group = "%s"
				  server_name                   = "%s"
				  partner_server_name           = "%s"
				  failover_group                = "%s"
				}`,
				azureTenantID,
				azureClientID,
				azureClientSecret,
				azureSubscriptionID,
				resourceGroup,
				partnerServerResourceGroup,
				serverName,
				partnerServerName,
				failoverGroup,
			)

			resource.Test(GinkgoT(), resource.TestCase{
				IsUnitTest: true, // means we don't need to set TF_ACC
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"csbmssqldbrunfailover": func() (*schema.Provider, error) { return csbmssqldbrunfailover.Provider(), nil },
				},
				Steps: []resource.TestStep{{
					ResourceName: "csbmssqldbrunfailover_failover",
					Config:       hcl,
					ExpectError:  regexp.MustCompile(expectError),
				}},
			})
		},
		Entry("tenant id", func() { azureTenantID = "not valid" }, `invalid value "not valid" for identifier "azure_tenant_id"`),
		Entry("client id", func() { azureClientID = "" }, `invalid value "" for identifier "azure_client_id"`),
		Entry("client secret", func() { azureClientSecret = "invalid value" }, `invalid value "invalid value" for identifier "azure_client_secret"`),
		Entry("subscription id", func() { azureSubscriptionID = "&&" }, `invalid value "&&" for identifier "azure_subscription_id"`),
		Entry("resource group", func() { resourceGroup = "not valid" }, `invalid value "not valid" for identifier "resource_group"`),
		Entry("partner server resource group", func() { partnerServerResourceGroup = "not valid" }, `invalid value "not valid" for identifier "partner_server_resource_group"`),
		Entry("server name", func() { serverName = "&&" }, `invalid value "&&" for identifier "server_name"`),
		Entry("partner server name", func() { partnerServerName = "&&" }, `invalid value "&&" for identifier "partner_server_name"`),
		Entry("failover group", func() { failoverGroup = "&&" }, `invalid value "&&" for identifier "failover_group"`),
	)
})
