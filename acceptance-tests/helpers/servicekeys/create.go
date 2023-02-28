package servicekeys

import (
	"time"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"
	"csbbrokerpakazure/acceptance-tests/helpers/random"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const timeout = 10 * time.Minute

type ServiceKey struct {
	name                string
	serviceInstanceName string
}

func Create(serviceInstanceName string) *ServiceKey {
	name := random.Name()
	session := cf.Start("create-service-key", serviceInstanceName, name)
	Eventually(session).WithTimeout(timeout).Should(Exit(0))

	return &ServiceKey{
		name:                name,
		serviceInstanceName: serviceInstanceName,
	}
}
