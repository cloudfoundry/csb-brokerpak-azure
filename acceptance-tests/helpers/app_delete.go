package helpers

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func AppDelete(names ...string) {
	for _, name := range names {
		session := StartCF("delete", "-f", name)
		Eventually(session, time.Minute).Should(Exit(0))
	}
}
