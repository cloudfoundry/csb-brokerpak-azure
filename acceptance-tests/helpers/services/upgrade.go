package services

import (
	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func (s *ServiceInstance) Upgrade() {
	var command []string
	switch cf.Version() {
	case cf.VersionV8:
		command = []string{"upgrade-service", s.Name, "--force"}
	default:
		command = []string{"update-service", s.Name, "--upgrade", "--force"}
	}

	session := cf.Start(command...)
	Eventually(session).WithTimeout(asyncCommandTimeout).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", s.Name)
		Expect(out).NotTo(MatchRegexp(`status:\s+update failed`))
		return out
	}).WithTimeout(operationTimeout).WithPolling(pollingInterval).Should(MatchRegexp(`status:\s+update succeeded`))
}
