package acceptance_test

import (
	"flag"
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
	metadata        environment.Metadata
	subscriptionID  string
	firewallStartIP string
	firewallEndIP   string
)

func init() {
	flag.StringVar(&firewallStartIP, "firewall-start-ip", "", "start IP for firewall hole")
	flag.StringVar(&firewallEndIP, "firewall-end-ip", "", "end IP for firewall hole")
	if firewallStartIP != "" && firewallEndIP == "" || firewallStartIP == "" && firewallEndIP != "" {
		panic("--firewall-start-ip and --firewall-end-ip must be specified together")
	}
}

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
