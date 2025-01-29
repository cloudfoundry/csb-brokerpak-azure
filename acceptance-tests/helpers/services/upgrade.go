package services

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"
)

func (s *ServiceInstance) Upgrade() {
	if !s.UpgradeAvailable() {
		GinkgoWriter.Println("No Upgrade available for service instance")
		return
	}

	session := cf.Start("upgrade-service", s.Name, "--force", "--wait")
	Eventually(session).WithTimeout(operationTimeout).Should(Exit(0), func() string {
		out, _ := cf.Run("service", s.Name)
		return out
	})

	out, _ := cf.Run("service", s.Name)
	Expect(out).To(MatchRegexp(`status:\s+update succeeded`))

	Expect(s.UpgradeAvailable()).To(BeFalse(), "service instance has an upgrade available after upgrade")
}

func (s *ServiceInstance) UpgradeAvailable() bool {
	out, _ := cf.Run("curl", fmt.Sprintf("/v3/service_instances/%s", s.GUID()))

	var receiver struct {
		UpgradeAvailable bool `json:"upgrade_available"`
	}
	Expect(json.Unmarshal([]byte(out), &receiver)).NotTo(HaveOccurred())
	return receiver.UpgradeAvailable
}

func (s *ServiceInstance) UpgradeExpectFailure() {
	Expect(s.UpgradeAvailable()).To(BeTrue(), "service instance does not have an upgrade available")

	session := cf.Start("upgrade-service", s.Name, "--force", "--wait")
	Eventually(session).WithTimeout(operationTimeout).Should(Exit())

	out, _ := cf.Run("service", s.Name)
	Expect(out).To(MatchRegexp(`status:\s+update failed`))
	Expect(out).To(MatchRegexp(`message:\s+upgrade failed: Error: A resource with the ID .* already exists`))
}
