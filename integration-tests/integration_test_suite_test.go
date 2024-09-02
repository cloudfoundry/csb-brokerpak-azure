package integration_test

import (
	"encoding/json"
	"strings"
	"testing"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTests Suite")
}

const (
	armClientID       = "arm-client-id"
	armClientSecret   = "arm-client-secret"
	armSubscriptionID = "arm-subscription-id"
	armTenantID       = "arm-tenant-id"
	Name              = "Name"
	ID                = "ID"
)

var (
	mockTerraform testframework.TerraformMock
	broker        *testframework.TestInstance
)

var _ = BeforeSuite(func() {
	var err error
	mockTerraform, err = testframework.NewTerraformMock()
	Expect(err).NotTo(HaveOccurred())

	broker, err = testframework.BuildTestInstance(
		testframework.PathToBrokerPack(),
		mockTerraform,
		GinkgoWriter,
		"service-images",
	)
	Expect(err).NotTo(HaveOccurred())
	servers := map[string]map[string]any{"preprovisioned-server-name": {"admin_password": "something"}}
	serverPairs := map[string]string{"preprovisioned-server-name": "another_server"}

	Expect(broker.Start(GinkgoWriter, []string{
		"GSB_COMPATIBILITY_ENABLE_PREVIEW_SERVICES=true",
		"ARM_CLIENT_ID=" + armClientID,
		"ARM_CLIENT_SECRET=" + armClientSecret,
		"ARM_SUBSCRIPTION_ID=" + armSubscriptionID,
		"ARM_TENANT_ID=" + armTenantID,
		"CSB_LISTENER_HOST=localhost",
		"GSB_SERVICE_CSB_AZURE_MSSQL_DB_PLANS=" + marshall(customMSSQLDBPlans),
		"MSSQL_DB_SERVER_CREDS=" + marshall(servers),
		"MSSQL_DB_FOG_SERVER_PAIR_CREDS=" + marshall(serverPairs),
		"GSB_COMPATIBILITY_ENABLE_GCP_DEPRECATED_SERVICES=true",
	})).To(Succeed())

})

var _ = AfterSuite(func() {
	if broker != nil {
		Expect(broker.Cleanup()).To(Succeed())
	}
})

func stringOfLen(length int) string {
	return strings.Repeat("a", length)
}

func marshall(element any) string {
	b, err := json.Marshal(element)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}
