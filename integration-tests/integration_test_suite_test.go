package integration_test

import (
	"encoding/json"
	"testing"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegrationTests Suite")
}

const (
	armClientId       = "arm-client-id"
	armClientSecret   = "arm-client-secret"
	armSubscriptionId = "arm-subscription-id"
	armTenantId       = "arm-tenant-id"
)

var (
	mockTerraform     testframework.TerraformMock
	broker            *testframework.TestInstance
	provisionDefaults = map[string]any{
		"location": "az",
	}
)

var _ = BeforeSuite(func() {
	var err error
	mockTerraform, err = testframework.NewTerraformMock()
	Expect(err).NotTo(HaveOccurred())

	extraFoldersBrokerpak := []string{"tools"}
	broker, err = testframework.BuildTestInstance(
		testframework.PathToBrokerPack(),
		mockTerraform,
		GinkgoWriter,
		extraFoldersBrokerpak...,
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(broker.Start(GinkgoWriter, []string{
		"GSB_COMPATIBILITY_ENABLE_PREVIEW_SERVICES=true",
		"ARM_CLIENT_ID=" + armClientId,
		"ARM_CLIENT_SECRET=" + armClientSecret,
		"ARM_SUBSCRIPTION_ID=" + armSubscriptionId,
		"ARM_TENANT_ID=" + armTenantId,
		"CSB_LISTENER_HOST=localhost",
	})).To(Succeed())
})

var _ = AfterSuite(func() {
	if broker != nil {
		Expect(broker.Cleanup()).To(Succeed())
	}
})

func marshall(element any) string {
	b, err := json.Marshal(element)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}
