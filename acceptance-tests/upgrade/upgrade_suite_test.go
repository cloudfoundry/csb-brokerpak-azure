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

	preflight(developmentBuildDir) // faster feedback as no download
	preflight(releasedBuildDir)
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
