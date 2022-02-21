package bindings

import (
	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
	gexec "github.com/onsi/gomega/gexec"
)

func (b *Binding) Unbind() {
	session := cf.Start("unbind-service", b.appName, b.serviceInstanceName)
	Eventually(session).WithTimeout(timeout).Should(gexec.Exit(0))
}
