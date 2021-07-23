package helpers

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func AppStart(names ...string) {
	for _, name := range names {
		session := StartCF("start", name)
		Eventually(session, 5*time.Minute).Should(Exit())

		if session.ExitCode() != 0 {
			fmt.Fprintf(GinkgoWriter, "FAILED to start app. Getting logs...")
			CF("logs", name, "--recent")
			Fail("App failed to start")
		}
	}
}
