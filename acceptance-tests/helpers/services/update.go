package services

import (
	"acceptancetests/helpers/cf"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func (s *ServiceInstance) Update(parameters ...string) {
	switch cf.Version() {
	case cf.VersionV8:
		s.updateServiceWithWait(parameters...)
	default:
		s.updateServiceWithPoll(parameters...)
	}
}

func (s *ServiceInstance) updateServiceWithWait(parameters ...string) {
	args := append([]string{"update-service", s.Name, "--wait"}, parameters...)

	session := cf.Start(args...)
	Eventually(session, time.Hour).Should(Exit(0), func() string {
		out, _ := cf.Run("service", s.Name)
		return out
	})
}

func (s *ServiceInstance) updateServiceWithPoll(parameters ...string) {
	args := append([]string{"update-service", s.Name}, parameters...)

	session := cf.Start(args...)
	Eventually(session, 5*time.Minute).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", s.Name)
		Expect(out).NotTo(MatchRegexp(`status:\s+update failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+update succeeded`))
}
