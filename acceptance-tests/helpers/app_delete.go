package helpers

import (
	"acceptancetests/helpers/cf"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func AppDelete(apps ...AppInstance) {
	for _, app := range apps {
		session := cf.Start("delete", "-f", app.name)
		Eventually(session, time.Minute).Should(Exit(0))
	}
}
