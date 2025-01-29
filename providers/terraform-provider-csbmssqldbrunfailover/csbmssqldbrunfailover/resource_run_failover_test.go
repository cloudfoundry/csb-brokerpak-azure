package csbmssqldbrunfailover_test

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbmssqldbrunfailover/csbmssqldbrunfailover"
	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbmssqldbrunfailover/testhelpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resource_run_failover resource", Ordered, Label("acceptance"), func() {

	var (
		failoverData        testhelpers.FailoverData
		azureTenantID       string
		azureClientID       string
		azureClientSecret   string
		azureSubscriptionID string
		config              testhelpers.FailoverConfig
	)

	BeforeAll(func() {
		azureSubscriptionID = os.Getenv("ARM_SUBSCRIPTION_ID")
		azureTenantID = os.Getenv("ARM_TENANT_ID")
		azureClientID = os.Getenv("ARM_CLIENT_ID")
		azureClientSecret = os.Getenv("ARM_CLIENT_SECRET")
		Expect(azureSubscriptionID).NotTo(BeEmpty(), "ARM_SUBSCRIPTION_ID environment variable should not be empty")
		Expect(azureTenantID).NotTo(BeEmpty(), "ARM_TENANT_ID environment variable should not be empty")
		Expect(azureClientID).NotTo(BeEmpty(), "ARM_CLIENT_ID environment variable should not be empty")
		Expect(azureClientSecret).NotTo(BeEmpty(), "ARM_CLIENT_SECRET environment variable should not be empty")
		Expect(os.Getenv("TF_ACC")).NotTo(BeEmpty(), "TF_ACC environment variable should not be empty")

		_ = os.Setenv("AZURE_SUBSCRIPTION_ID", azureSubscriptionID)
		_ = os.Setenv("AZURE_TENANT_ID", azureTenantID)
		_ = os.Setenv("AZURE_CLIENT_ID", azureClientID)
		_ = os.Setenv("AZURE_CLIENT_SECRET", azureClientSecret)
		var err error

		config = testhelpers.FailoverConfig{
			ResourceGroupName:     testhelpers.RandomName("resourcegroupname"),
			ServerName:            testhelpers.RandomName("servername"),
			MainLocation:          "eastus",
			PartnerServerLocation: "eastus2",
			PartnerServerName:     testhelpers.RandomName("partnerservername"),
			SubscriptionID:        azureSubscriptionID,
			FailoverGroupName:     testhelpers.RandomName("failovergroupname"),
		}

		failoverData, err = testhelpers.CreateFailoverGroup(config)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			_ = testhelpers.Cleanup(config)
		})

	})

	It("should failover to the secondary database", func() {
		hcl := generateHCLContent(
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
