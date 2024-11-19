package upgrade_test

import (
	"flag"
	"os"
	"testing"

	"csbbrokerpakazure/acceptance-tests/helpers/brokerpaks"
	"csbbrokerpakazure/acceptance-tests/helpers/environment"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	fromVersion           string
	developmentBuildDir   string
	releasedBuildDir      string
	intermediateBuildDirs string
	metadata              environment.Metadata
	subscriptionID        string
)

func init() {
	flag.StringVar(&fromVersion, "from-version", "", "version to upgrade from")
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../..", "location of development version of built broker and brokerpak")
	flag.StringVar(&intermediateBuildDirs, "intermediateBuildDirs", "", "comma separated locations of intermediate versions of built broker and brokerpak")
}

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrade Suite")
}

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()

	if releasedBuildDir == "" { // Released dir not specified, so we should download a brokerpak
		if fromVersion == "" { // Version not specified, so use latest
			fromVersion = brokerpaks.LatestVersion()
		}

		releasedBuildDir = brokerpaks.DownloadBrokerpak(fromVersion, brokerpaks.TargetDir(fromVersion))
	}

	preflight(developmentBuildDir)
	// Don't do a preflight on releasedBuildDir as older versions (tile v1.3.0 and earlier) don't have a .envrc file

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

// preflight checks that a specified broker dir is viable so that the user gets fast feedback
func preflight(dir string) {
	GinkgoHelper()

	entries, err := os.ReadDir(dir)
	Expect(err).NotTo(HaveOccurred())
	names := make([]string, len(entries))
	for i := range entries {
		names[i] = entries[i].Name()
	}

	Expect(names).To(ContainElements(
		Equal("cloud-service-broker"),
		Equal(".envrc"),
		MatchRegexp(`azure-services-\S+\.brokerpak`),
	))
}
