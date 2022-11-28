package integration_test

import (
	"strings"
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
	armClientID       = "arm-client-id"
	armClientSecret   = "arm-client-secret"
	armSubscriptionID = "arm-subscription-id"
	armTenantID       = "arm-tenant-id"
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
	Expect(broker.Start(GinkgoWriter, []string{
		"GSB_COMPATIBILITY_ENABLE_PREVIEW_SERVICES=true",
		"ARM_CLIENT_ID=" + armClientID,
		"ARM_CLIENT_SECRET=" + armClientSecret,
		"ARM_SUBSCRIPTION_ID=" + armSubscriptionID,
		"ARM_TENANT_ID=" + armTenantID,
		"CSB_LISTENER_HOST=localhost",
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
