package upgrade_test

import (
	"flag"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega/gexec"

	"csbbrokerpakazure/acceptance-tests/helpers/environment"

	. "github.com/onsi/ginkgo/v2"
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

func fetchResourceID(kind, name, server string) string {
	command := exec.Command("az", "sql", kind, "show", "--name", name, "--server", server, "--resource-group", metadata.ResourceGroup, "--query", "id", "-o", "tsv")
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, time.Minute).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}
