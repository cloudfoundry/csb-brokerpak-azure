package upgrade_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var developmentBuildDir string
var releasedBuildDir string
var brokerName string

func init() {
	flag.StringVar(&releasedBuildDir, "releasedBuildDir", "../../../azure-released", "location of released version of built broker and brokerpak")
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../../dev-release", "location of development version of built broker and brokerpak")
	brokerName = helpers.RandomName("csb")
}

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrade Suite")
}