package upgrade_test

import (
	"flag"
	"strings"
	"testing"

	"csbbrokerpakazure/acceptance-tests/helpers/environment"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	developmentBuildDir   string
	releasedBuildDir      string
	intermediateBuildDirs []string
	metadata              environment.Metadata
)

func init() {
	var intermediateBuildDirsFlag string
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "../../../azure-released", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../../dev-release", "location of development version of built broker and brokerpak")
	flag.StringVar(&intermediateBuildDirsFlag, "intermediateBuildDirs", "", "comma separated locations of intermediate versions of built broker and brokerpak")

	intermediateBuildDirs = strings.Split(intermediateBuildDirsFlag, ",")
}

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
})

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrade Suite")
}
