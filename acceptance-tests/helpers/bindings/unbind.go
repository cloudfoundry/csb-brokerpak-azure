package bindings

import (
	"acceptancetests/helpers/cf"

	. "github.com/onsi/gomega"
	gexec "github.com/onsi/gomega/gexec"
)

func (b *Binding) Unbind() {
	session := cf.Start("unbind-service", b.appName, b.serviceInstanceName)
	Eventually(session).WithTimeout(timeout).Should(gexec.Exit(0))
}
