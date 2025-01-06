package services

import (
	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func (s *ServiceInstance) Delete() {
	Delete(s.Name)
}

func Delete(name string) {
	switch cf.Version() {
	case cf.VersionV8:
		deleteWithWait(name)
	default:
		deleteWithPoll(name)
	}
}

func deleteWithWait(name string) {
	session := cf.Start("delete-service", "-f", name, "--wait")
	Eventually(session).WithTimeout(operationTimeout).Should(Exit(0))
}

func deleteWithPoll(name string) {
	cf.Run("delete-service", "-f", name)

	Eventually(func() string {
		out, _ := cf.Run("services")
		return out
	}).WithTimeout(operationTimeout).WithPolling(pollingInterval).ShouldNot(ContainSubstring(name))
}

func (s *ServiceInstance) Purge() {
	cf.Run("purge-service-instance", "-f", s.Name)
}
