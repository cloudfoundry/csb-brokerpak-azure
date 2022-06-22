// Package servicekeys manages service keys
package servicekeys

import (
	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func (s *ServiceKey) Delete() {
	session := cf.Start("delete-service-key", "-f", s.serviceInstanceName, s.name)
	Eventually(session).WithTimeout(timeout).Should(Exit(0))
}
