package csbmssqldbrunfailover_test

import (
	"fmt"

	"csbbrokerpakazure/providers/terraform-provider-csbmssqldbrunfailover/csbmssqldbrunfailover"
	"csbbrokerpakazure/providers/terraform-provider-csbmssqldbrunfailover/testhelpers"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/dsl/core"
)

var _ = Describe("resource_run_failover resource", func() {

	var (
		// failoverData        testhelpers.FailoverData
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

		// failoverData = testhelpers.CreateFailoverGroup(testhelpers.FailoverConfig{
		// 	ResourceGroupName: "resourcegroupname2",
		// 	ServerName:        "servername2",
		// 	Location:          "eastus",
		// 	PartnerServerName: "partnerservername2",
		// 	SubscriptionID:    azureSubscriptionID,
		// 	FailoverGroupName: "failovergroupname2",
		// })

	})

	It("should initiate the failover operation", func() {
		hcl := fmt.Sprintf(`
				provider "csbmssqldbrunfailover" {
				  azure_tenant_id       = "%s"
				  azure_client_id       = "%s"
				  azure_client_secret   = "%s"
				  azure_subscription_id = "%s"
				}
				
				resource "csbmssqldbrunfailover_failover" "failover" {
				  resource_group      = "%s"
				  server_name         = "%s"
				  partner_server_name = "%s"	
				  failover_group      = "%s"
				}`,
			azureTenantID,
			azureClientID,
			azureClientSecret,
			azureSubscriptionID,
			// *failoverData.ResourceGroup.Name,
			// *failoverData.Server.Name,
			// *failoverData.FailoverGroup.Name,
			"resourcegroupname2",
			"servername2",
			"partnerservername2",
			"failovergroupname2",
		)

		resource.Test(GinkgoT(), resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"csbmssqldbrunfailover": func() (*schema.Provider, error) { return csbmssqldbrunfailover.Provider(), nil },
			},
			Steps: []resource.TestStep{{
				ResourceName: "csbmssqldbrunfailover_failover",
				Config:       hcl,
				Check: func(state *terraform.State) error {
					// group, err := testhelpers.GetFailoverGroup(
					// 	*failoverData.ResourceGroup.Name,
					// 	*failoverData.Server.Name,
					// 	*failoverData.FailoverGroup.Name,
					// 	azureSubscriptionID,
					// )
					group, err := testhelpers.GetFailoverGroup(
						"resourcegroupname2",
						"servername2",
						"failovergroupname2",
						azureSubscriptionID,
					)
					if err != nil {
						return fmt.Errorf("error getting failover group %w", err)
					}

					partnerInfo := group.Properties.PartnerServers[0]
					got := *partnerInfo.ReplicationRole
					want := armsql.FailoverGroupReplicationRolePrimary
					if got != want {
						return fmt.Errorf("failover group replication role error got %s - want %s", got, want)
					}

					return nil
				},
			}},
			CheckDestroy: func(state *terraform.State) error {
				core.GinkgoWriter.Printf("state %+v", state)
				core.GinkgoWriter.Println()
				// group, err := testhelpers.GetFailoverGroup(
				// 	*failoverData.ResourceGroup.Name,
				// 	*failoverData.Server.Name,
				// 	*failoverData.FailoverGroup.Name,
				// 	azureSubscriptionID,
				// )
				group, err := testhelpers.GetFailoverGroup(
					"resourcegroupname2",
					"servername2",
					"failovergroupname2",
					azureSubscriptionID,
				)
				if err != nil {
					return fmt.Errorf("error getting failover group %w", err)
				}

				partnerInfo := group.Properties.PartnerServers[0]
				got := *partnerInfo.ReplicationRole
				want := armsql.FailoverGroupReplicationRoleSecondary
				if got != want {
					return fmt.Errorf("failover group replication role error got %s - want %s", got, want)
				}

				return nil
			},
		})
	})
})

// testAccPreCheck validates the necessary test API keys exist
// in the testing environment
func testAccPreCheck() {
	// if v := os.Getenv("EXAMPLE_KEY"); v == "" {
	// 	t.Fatal("EXAMPLE_KEY must be set for acceptance tests")
	// }
	// Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
	//
	// if v := os.Getenv("EXAMPLE_SECRET"); v == "" {
	// 	t.Fatal("EXAMPLE_SECRET must be set for acceptance tests")
	// }
}
