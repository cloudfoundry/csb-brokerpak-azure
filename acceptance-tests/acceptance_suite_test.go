package acceptance_test

import (
	"os"
	"testing"

	"csbbrokerpakazure/acceptance-tests/helpers/environment"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAcceptanceTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Tests Suite")
}

var (
	metadata       environment.Metadata
	subscriptionID string
)

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
	subscriptionID = os.Getenv("ARM_SUBSCRIPTION_ID")
	Expect(subscriptionID).NotTo(BeEmpty(), "ARM_SUBSCRIPTION_ID environment variable should not be empty")
	Expect(os.Getenv("ARM_TENANT_ID")).NotTo(BeEmpty(), "ARM_TENANT_ID environment variable should not be empty")
	Expect(os.Getenv("ARM_CLIENT_ID")).NotTo(BeEmpty(), "ARM_CLIENT_ID environment variable should not be empty")
	Expect(os.Getenv("ARM_CLIENT_SECRET")).NotTo(BeEmpty(), "ARM_CLIENT_SECRET environment variable should not be empty")

	_ = os.Setenv("AZURE_SUBSCRIPTION_ID", subscriptionID)
	_ = os.Setenv("AZURE_TENANT_ID", os.Getenv("ARM_TENANT_ID"))
	_ = os.Setenv("AZURE_CLIENT_ID", os.Getenv("ARM_CLIENT_ID"))
	_ = os.Setenv("AZURE_CLIENT_SECRET", os.Getenv("ARM_CLIENT_SECRET"))
})
