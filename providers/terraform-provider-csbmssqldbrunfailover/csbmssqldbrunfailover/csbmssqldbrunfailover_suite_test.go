package csbmssqldbrunfailover_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCSBSQLServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSB MSSQL Db Run Failover Terraform Provider Suite")
}

const (
	subscriptionIDKey = "subscriptionIDKey"
	tenantIDKey       = "tenantIDKey"
	clientIDKey       = "clientIDKey"
	clientSecretKey   = "clientSecretKey"
)

type azureCreds map[string]string

func (a azureCreds) getSubscriptionID() string {
	return a[subscriptionIDKey]
}
func (a azureCreds) getTenantID() string {
	return a[tenantIDKey]
}
func (a azureCreds) getClientID() string {
	return a[clientIDKey]
}
func (a azureCreds) getClientSecret() string {
	return a[clientSecretKey]
}

var creds = azureCreds{}

var _ = BeforeSuite(func() {
	creds[subscriptionIDKey] = os.Getenv("ARM_SUBSCRIPTION_ID")
	creds[tenantIDKey] = os.Getenv("ARM_TENANT_ID")
	creds[clientIDKey] = os.Getenv("ARM_CLIENT_ID")
	creds[clientSecretKey] = os.Getenv("ARM_CLIENT_SECRET")
	Expect(creds.getSubscriptionID()).NotTo(BeEmpty())

	_ = os.Setenv("AZURE_SUBSCRIPTION_ID", creds.getSubscriptionID())
	_ = os.Setenv("AZURE_TENANT_ID", creds.getTenantID())
	_ = os.Setenv("AZURE_CLIENT_ID", creds.getClientID())
	_ = os.Setenv("AZURE_CLIENT_SECRET", creds.getClientSecret())
})
