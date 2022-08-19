package csbmssqldbrunfailover_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCSBSQLDBRunFailoverServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSB MSSQL Db Run Failover Terraform Provider Suite")
}

func generateHCLContent(
	azureTenantID,
	azureClientID,
	azureClientSecret,
	azureSubscriptionID,
	resourceGroup,
	partnerServerResourceGroup,
	serverName,
	partnerServerName,
	failoverGroup string,
) string {
	const hcl = `
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
}`
	return fmt.Sprintf(
		hcl,
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
}
