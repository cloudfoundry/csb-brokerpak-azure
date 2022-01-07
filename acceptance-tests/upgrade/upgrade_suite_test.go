package upgrade_test

import (
	"acceptancetests/helpers/environment"
	"flag"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	developmentBuildDir string
	releasedBuildDir    string
	metadata            environment.Metadata
)

func init() {
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "../../../azure-released", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../../dev-release", "location of development version of built broker and brokerpak")
}

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
})

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrade Suite")
}
