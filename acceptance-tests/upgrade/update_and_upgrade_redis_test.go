package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeRedisTest", func() {
	Context("When upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			brokerName := helpers.RandomName("csb-redis")
			serviceBroker := helpers.PushAndStartBroker(brokerName, releasedBuildDir)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := helpers.CreateServiceFromBroker("csb-azure-redis", "small", brokerName)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.Redis)
			appTwo := helpers.AppPushUnstarted(apps.Redis)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding to the apps")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppStart(appOne, appTwo)

			By("setting a key-value using the first app")
			key1 := helpers.RandomHex()
			value1 := helpers.RandomHex()
			appOne.PUT(value1, key1)

			By("getting the value using the second app")
			got := appTwo.GET(key1)
			Expect(got).To(Equal(value1))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)
			key2 := helpers.RandomHex()
			value2 := helpers.RandomHex()
			appOne.PUT(value2, key2)
			Expect(appTwo.GET(key2)).To(Equal(value2))

			By("getting the value using the second app")
			Expect(appTwo.GET(key1)).To(Equal(value1))

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("checking it still works")
			key3 := helpers.RandomHex()
			value3 := helpers.RandomHex()
			appOne.PUT(value3, key3)
			Expect(appTwo.GET(key3)).To(Equal(value3))
		})
	})
})
