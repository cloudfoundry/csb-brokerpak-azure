package helpers

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func AppPushUnstarted(prefix, appDir string) string {
	name := RandomName(prefix)
	session := StartCF("push", "--no-start", "-b", "binary_buildpack", "-p", appDir, name)
	Eventually(session, 5*time.Minute).Should(Exit())

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to push app. Getting logs...")
		CF("logs", name, "--recent")
		Fail("App failed to push")
	}

	return name
}
