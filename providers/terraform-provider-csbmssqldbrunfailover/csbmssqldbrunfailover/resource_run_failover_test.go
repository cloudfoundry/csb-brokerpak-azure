package csbmssqldbrunfailover_test

import (
	"fmt"

	"csbbrokerpakazure/providers/terraform-provider-csbmssqldbrunfailover/csbmssqldbrunfailover"
	"csbbrokerpakazure/providers/terraform-provider-csbmssqldbrunfailover/testhelpers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("resource_run_failover resource", func() {

	var (
		failoverData        testhelpers.FailoverData
		azureTenantID       string
		azureClientID       string
		azureClientSecret   string
		azureSubscriptionID string
	)
	BeforeEach(func() {
		azureTenantID = creds.getTenantID()
		azureClientID = creds.getClientID()
		azureClientSecret = creds.getClientSecret()
		azureSubscriptionID = creds.getSubscriptionID()

		failoverData = testhelpers.CreateFailoverGroup(testhelpers.FailoverConfig{
			ResourceGroupName: "resourcegroupname",
			ServerName:        "servername",
			Location:          "eastus",
			PartnerServerName: "partnerservername",
			SubscriptionID:    azureSubscriptionID,
			FailoverGroupName: "failovergroupname",
		})

	})

	It("should initiate the failover operation", func() {
		fmt.Println(failoverData)
		hcl := fmt.Sprintf(`
				provider "csbmssqldbrunfailover" {
				  azure_tenant_id       = "%s"
				  azure_client_id       = "%s"
				  azure_client_secret   = "%s"
				  azure_subscription_id = "%s"
				}
				
				resource "csbmssqldbrunfailover_failover" "failover" {
				  resource_group = "%s"
				  server_name    = "%s"
				  failover_group = "%s"
				}`,
			azureTenantID,
			azureClientID,
			azureClientSecret,
			azureSubscriptionID,
			*failoverData.ResourceGroup.Name,
			*failoverData.Server.Name,
			*failoverData.FailoverGroup.Name,
		)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"csbmssqldbrunfailover": func() (*schema.Provider, error) { return csbmssqldbrunfailover.Provider(), nil },
			},
			Steps: []resource.TestStep{{
				ResourceName: "csbmssqldbrunfailover_failover",
				Config:       hcl,
				Check: func(state *terraform.State) error {
					group, err := testhelpers.GetFailoverGroup(
						*failoverData.ResourceGroup.Name,
						*failoverData.Server.Name,
						*failoverData.FailoverGroup.Name,
						azureSubscriptionID,
					)
					if err != nil {
						return fmt.Errorf("error getting failover group %w", err)
					}
					fmt.Printf("%+v", group)
					return fmt.Errorf("%+v", group)
				},
			}},
			CheckDestroy: func(state *terraform.State) error {
				fmt.Printf("state %+v", state)
				return nil
			},
		})
	})
})
