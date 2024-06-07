package csbmssqldbrunfailover_test

import (
	"regexp"

	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbmssqldbrunfailover/csbmssqldbrunfailover"
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

			hcl := generateHCLContent(
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
		Entry(
			"tenant id",
			func() { azureTenantID = "" },
			`empty value for identifier "azure_tenant_id"`,
		),
		Entry(
			"client id",
			func() { azureClientID = "" },
			`empty value for identifier "azure_client_id"`,
		),
		Entry(
			"client secret empty",
			func() { azureClientSecret = "" },
			`empty value for identifier "azure_client_secret"`,
		),
		Entry(
			"subscription id",
			func() { azureSubscriptionID = "" },
			`empty value for identifier "azure_subscription_id"`,
		),
		Entry(
			"resource group",
			func() { resourceGroup = "" },
			`empty value for identifier "resource_group"`,
		),
		Entry(
			"partner server resource group",
			func() { partnerServerResourceGroup = "" },
			`empty value for identifier "partner_server_resource_group"`,
		),
		Entry(
			"server name",
			func() { serverName = "" },
			`empty value for identifier "server_name"`,
		),
		Entry(
			"partner server name",
			func() { partnerServerName = "" },
			`empty value for identifier "partner_server_name"`,
		),
		Entry(
			"failover group",
			func() { failoverGroup = "" },
			`empty value for identifier "failover_group"`,
		),
	)
})
