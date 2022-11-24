package terraformtests

import (
	"os"
	"testing"

	"golang.org/x/exp/maps"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cp "github.com/otiai10/copy"
)

func TestTerraformTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TerraformTests Suite")
}

var (
	workingDir          string
	azureClientID       = os.Getenv("ARM_CLIENT_ID")
	azureClientSecret   = os.Getenv("ARM_CLIENT_SECRET")
	azureSubscriptionID = os.Getenv("ARM_SUBSCRIPTION_ID")
	azureTenantID       = os.Getenv("ARM_TENANT_ID")
)

var _ = BeforeSuite(func() {
	workingDir = GinkgoT().TempDir()
	Expect(cp.Copy("../terraform", workingDir)).NotTo(HaveOccurred())
})

func buildVars(defaults, overrides map[string]any) map[string]any {
	result := map[string]any{}
	maps.Copy(result, defaults)
	maps.Copy(result, overrides)
	return result
}
