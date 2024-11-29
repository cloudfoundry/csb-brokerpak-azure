package upgrade_test

import (
	"fmt"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/az"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/lookupplan"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func updateRedisFirewall(serviceName, resourceGroup, publicIP string) {
	az.Run("redis",
		"firewall-rules",
		"create",
		"--name", serviceName,
		"--resource-group", resourceGroup,
		"--rule-name", "allowtestrule",
		"--start-ip", publicIP,
		"--end-ip", publicIP,
	)
}

var _ = Describe("UpgradeRedisTest", Label("redis"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-redis"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a resource group")
			az.Start("group", "create", "--name", metadata.ResourceGroup, "--location", defaultRegion)

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-azure-redis",
				lookupplan.LookupByID("6b9ca24e-1dec-4e6f-8c8a-dc6e11ab5bef", "csb-azure-redis", serviceBroker.Name),
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("changing the firewall to allow comms")
			serviceName := fmt.Sprintf("csb-redis-%s", serviceInstance.GUID())
			updateRedisFirewall(serviceName, metadata.ResourceGroup, metadata.PublicIP)

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.Redis))
			appTwo := apps.Push(apps.WithApp(apps.Redis))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)
			apps.Start(appOne, appTwo)

			By("setting a key-value using the first app")
			key1 := random.Hexadecimal()
			value1 := random.Hexadecimal()
			appOne.PUT(value1, key1)

			By("getting the value using the second app")
			got := appTwo.GET(key1)
			Expect(got).To(Equal(value1))

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("changing the firewall to allow comms")
			updateRedisFirewall(serviceName, metadata.ResourceGroup, metadata.PublicIP)

			By("checking previously written data still accessible")
			Expect(appTwo.GET(key1)).To(Equal(value1))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)
			key2 := random.Hexadecimal()
			value2 := random.Hexadecimal()
			appOne.PUT(value2, key2)
			Expect(appTwo.GET(key2)).To(Equal(value2))

			By("checking previously written data still accessible")
			Expect(appTwo.GET(key1)).To(Equal(value1))

			By("updating the instance plan")
			serviceInstance.Update("-c", `{}`)

			By("changing the firewall to allow comms")
			updateRedisFirewall(serviceName, metadata.ResourceGroup, metadata.PublicIP)

			By("checking it still works")
			key3 := random.Hexadecimal()
			value3 := random.Hexadecimal()
			appOne.PUT(value3, key3)
			Expect(appTwo.GET(key3)).To(Equal(value3))
		})
	})
})
