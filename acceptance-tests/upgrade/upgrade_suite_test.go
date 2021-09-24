package upgrade_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var developmentBuildDir string

func init() {
	flag.StringVar(&developmentBuildDir, "developmentBuildDir", "/test-broker-update", "location of built broker and brokerpak")
}

func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrade Suite")
}
