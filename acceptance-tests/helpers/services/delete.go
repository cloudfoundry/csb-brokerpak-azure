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
	session := cf.Start("delete-service", "-f", name, "--wait")
	Eventually(session).WithTimeout(operationTimeout).Should(Exit(0))
}

func (s *ServiceInstance) Purge() {
	cf.Run("purge-service-instance", "-f", s.Name)
}
