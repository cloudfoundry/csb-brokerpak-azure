package helpers

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func AppStart(apps ...AppInstance) {
	for _, app := range apps {
		session := StartCF("start", app.name)
		Eventually(session, 5*time.Minute).Should(Exit())

		if session.ExitCode() != 0 {
			fmt.Fprintf(GinkgoWriter, "FAILED to start app. Getting logs...")
			CF("logs", app.name, "--recent")
			Fail("App failed to start")
		}
	}
}

func AppRestage(apps ...AppInstance) {
	for _, app := range apps {
		session := StartCF("restage", app.name)
		Eventually(session, 5*time.Minute).Should(Exit())

		if session.ExitCode() != 0 {
			fmt.Fprintf(GinkgoWriter, "FAILED to restage app. Getting logs...")
			CF("logs", app.name, "--recent")
			Fail("App failed to start")
		}
	}
}
