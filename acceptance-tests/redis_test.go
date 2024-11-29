package acceptance_test

import (
	"fmt"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/az"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Tests the *csb-azure-redis* service offering
// Uses the *default broker*
var _ = Describe("Redis", Label("redis"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-azure-redis", "deprecated-small")
		defer serviceInstance.Delete()

		By("updating the firewall to allow comms")
		serviceName := fmt.Sprintf("csb-redis-%s", serviceInstance.GUID())
		az.Run("redis",
			"firewall-rules",
			"create",
			"--name", serviceName,
			"--resource-group", metadata.ResourceGroup,
			"--rule-name", "allowtestrule",
			"--start-ip", metadata.PublicIP,
			"--end-ip", metadata.PublicIP,
		)

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.Redis))
		appTwo := apps.Push(apps.WithApp(apps.Redis))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the Redis service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appOne.PUT(value, key)

		By("getting the value using the second app")
		got := appTwo.GET(key)
		Expect(got).To(Equal(value))
	})
})
