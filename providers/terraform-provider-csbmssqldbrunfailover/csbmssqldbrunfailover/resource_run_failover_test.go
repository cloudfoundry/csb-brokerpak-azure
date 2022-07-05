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
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

var _ = Describe("resource_run_failover resource", Ordered, func() {

	var (
		failoverData        testhelpers.FailoverData
		azureTenantID       string
		azureClientID       string
		azureClientSecret   string
		azureSubscriptionID string
		config              testhelpers.FailoverConfig
	)

	BeforeAll(func() {
		var err error
		azureTenantID = creds.getTenantID()
		azureClientID = creds.getClientID()
		azureClientSecret = creds.getClientSecret()
		azureSubscriptionID = creds.getSubscriptionID()

		config = testhelpers.FailoverConfig{
			ResourceGroupName:     fmt.Sprintf("resourcegroupname-%s", uuid.New()),
			ServerName:            fmt.Sprintf("servername-%s", uuid.New()),
			MainLocation:          "eastus",
			PartnerServerLocation: "eastus2",
			PartnerServerName:     fmt.Sprintf("partnerservername-%s", uuid.New()),
			SubscriptionID:        azureSubscriptionID,
			FailoverGroupName:     fmt.Sprintf("failovergroupname-%s", uuid.New()),
		}

		failoverData, err = testhelpers.CreateFailoverGroup(config)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			_ = testhelpers.Cleanup(config)
		})

	})

	It("should failover to the secondary database", func() {
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
			*failoverData.ResourceGroup.Name,
			*failoverData.PartnerServerResourceGroup.Name,
			*failoverData.Server.Name,
			*failoverData.PartnerServer.Name,
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
					GinkgoWriter.Println("*******************Check Step*******************")
					group, err := testhelpers.GetFailoverGroup(
						*failoverData.ResourceGroup.Name,
						*failoverData.Server.Name,
						*failoverData.FailoverGroup.Name,
						azureSubscriptionID,
					)
					if err != nil {
						return fmt.Errorf("error getting failover group %w", err)
					}

					partnerInfo := group.Properties.PartnerServers[0]
					got := *partnerInfo.ReplicationRole
					want := armsql.FailoverGroupReplicationRolePrimary
					if got != want {
						return fmt.Errorf("failover group replication role error in create got %s - want %s", got, want)
					}

					return nil
				},
			}},
			CheckDestroy: func(state *terraform.State) error {
				GinkgoWriter.Println("*******************Check Destroy Step*******************")
				group, err := testhelpers.GetFailoverGroup(
					*failoverData.ResourceGroup.Name,
					*failoverData.Server.Name,
					*failoverData.FailoverGroup.Name,
					azureSubscriptionID,
				)
				if err != nil {
					return fmt.Errorf("error getting failover group %w", err)
				}

				partnerInfo := group.Properties.PartnerServers[0]
				got := *partnerInfo.ReplicationRole
				want := armsql.FailoverGroupReplicationRoleSecondary
				if got != want {
					return fmt.Errorf("failover group replication role error in destroy got %s - want %s", got, want)
				}

				return nil
			},
		})
	})
})
