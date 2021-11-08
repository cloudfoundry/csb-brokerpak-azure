package upgrade_test

import (
	"code.cloudfoundry.org/jsonry"
	"flag"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var developmentBuildDir string
var releasedBuildDir string

func init() {
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "/Users/normanja/workspace/csb/azure-released", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "/Users/normanja/workspace/csb/test-broker-update", "location of development version of built broker and brokerpak")
}

var _ = BeforeSuite(func() {
	file := os.Getenv("ENVIRONMENT_LOCK_METADATA")
	Expect(file).NotTo(BeEmpty(), "You must set the ENVIRONMENT_LOCK_METADATA environment variable")

	contents, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())

	Expect(jsonry.Unmarshal(contents, &metadata)).NotTo(HaveOccurred())
	Expect(metadata.ResourceGroup).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedSQLUsername).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedSQLPassword).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedSQLServer).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedSQLLocation).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedFOGUsername).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedFOGPassword).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedFOGServer).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedFOGLocation).NotTo(BeEmpty())
})

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrade Suite")
}