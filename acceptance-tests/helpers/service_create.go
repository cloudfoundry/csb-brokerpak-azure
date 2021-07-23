package helpers

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func CreateService(offering, plan, name string, parameters ...string) {
	session := StartCreateService(offering, plan, name, parameters...)
	Eventually(session, time.Minute).Should(Exit(0))
	Eventually(func() string {
		out, _ := CF("service", name)
		return out
	}, 30*time.Minute, 30*time.Second).Should(MatchRegexp(`status:\s+create succeeded`))
}

func StartCreateService(offering, plan, name string, parameters ...string) *Session {
	args := []string{"create-service", offering, plan, name}
	if len(parameters) > 0 {
		args = append(args, "-c", parameters[0])
	}

	session := StartCF(args...)
	return session
}
