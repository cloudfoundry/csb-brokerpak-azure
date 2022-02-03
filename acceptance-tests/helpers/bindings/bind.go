package bindings

import (
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const timeout = 10 * time.Minute

type Binding struct {
	name                string
	serviceInstanceName string
	appName             string
}

func Bind(serviceInstanceName, appName string) *Binding {
	name := random.Name()
	session := cf.Start("bind-service", appName, serviceInstanceName, "--binding-name", name)
	Eventually(session).WithTimeout(timeout).Should(Exit(0))
	return &Binding{
		name:                name,
		serviceInstanceName: serviceInstanceName,
		appName:             appName,
	}
}
