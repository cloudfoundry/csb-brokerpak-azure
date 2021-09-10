package errormessages_test

import (
	"acceptancetests/helpers"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Error Messages", func() {
	When("the create-service command fails immediately", func() {
		It("prints a useful error message", func() {
			name := helpers.RandomName("error")
			defer helpers.CF("delete-service", "-f", name)

			session := helpers.StartCF("create-service", "csb-azure-mysql", "small", name, "-c", `{"location":"bogus"}`)
			Eventually(session, time.Minute).Should(Exit(1))
			Expect(session.Out).To(Say(`FAILED\n`))
			Expect(session.Err).To(Say(`Service broker error: 1 error\(s\) occurred:.*location: location must be one of the following:( "\S+",?)+\n$`))
		})
	})

	When("the service creation fail asynchronously", func() {
		It("puts a useful error message in the service description", func() {
			name := helpers.RandomName("error")
			defer helpers.CF("delete-service", "-f", name)

			session := helpers.StartCF("create-service", "csb-azure-storage-account", "standard", name, "-c", `{"resource_group":"bogus"}`)
			Eventually(session, time.Minute).Should(Exit(0))

			Eventually(func() string {
				stdout, _ := helpers.CF("service", name)
				return stdout
			}, 10*time.Minute, 10*time.Second).Should(MatchRegexp(`status:\s+create failed`))

			stdout, _ := helpers.CF("service", name)
			Expect(stdout).To(MatchRegexp(`message:\s+Error: Error creating Azure Storage Account "\S+":.*Original Error: Code="ResourceGroupNotFound" Message="Resource group 'bogus' could not be found."`))
		})
	})
})
