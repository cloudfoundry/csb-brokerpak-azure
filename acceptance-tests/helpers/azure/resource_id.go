package azure

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"strings"
	"time"
)

func FetchResourceID(kind, name, server, resourceGroup string) string {
	command := exec.Command("az", "sql", kind, "show", "--name", name, "--server", server, "--resource-group", resourceGroup, "--query", "id", "-o", "tsv")
	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(session, time.Minute).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}
