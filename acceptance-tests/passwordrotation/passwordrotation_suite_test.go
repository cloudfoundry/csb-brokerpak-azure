package passwordrotation_test

import (
	"acceptancetests/helpers"
	"flag"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var developmentBuildDir string
var brokerName string

func init() {
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "../../dev-release", "location of development version of built broker and brokerpak")
	brokerName = helpers.RandomName("csb")
}

func TestKeyrotation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Password Rotation Suite")
}
