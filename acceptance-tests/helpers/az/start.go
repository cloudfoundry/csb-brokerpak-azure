// Package az to wrap executing the az cli
package az

import (
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func Run(args ...string) {
	GinkgoHelper()

	GinkgoWriter.Printf("Running: az %s\n", strings.Join(args, " "))
	command := exec.Command("az", args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).WithTimeout(time.Hour).Should(gexec.Exit(0))
}
