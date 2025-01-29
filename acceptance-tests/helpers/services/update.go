package services

import (
	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func (s *ServiceInstance) Update(parameters ...string) {
	args := append([]string{"update-service", s.Name, "--wait"}, parameters...)

	session := cf.Start(args...)
	Eventually(session).WithTimeout(operationTimeout).Should(Exit(0), func() string {
		out, _ := cf.Run("service", s.Name)
		return out
	})
}
