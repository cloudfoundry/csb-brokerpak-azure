package helpers

import (
	"os/exec"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func StartCF(args ...string) *Session {
	command := exec.Command("cf", args...)
	session, err := Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

func CF(args ...string) (string, string) {
	session := StartCF(args...)
	Eventually(session, time.Minute).Should(Exit(0))
	return string(session.Out.Contents()), string(session.Err.Contents())
}
