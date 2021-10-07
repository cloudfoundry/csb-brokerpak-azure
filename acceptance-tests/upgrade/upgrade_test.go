package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeTest", func() {
	Context("When upgrading broker version", func(){
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := helpers.PushAndStartBroker(brokerName, releasedBuildDir)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := helpers.CreateServiceInBroker("csb-azure-redis", "small", brokerName)
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
			key := helpers.RandomHex()
			value := helpers.RandomHex()
			appOne.PUT(value, key)

			By("getting the value using the second app")
			got := appTwo.GET(key)
			Expect(got).To(Equal(value))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)
			key = helpers.RandomHex()
			value = helpers.RandomHex()
			appOne.PUT(value, key)
			got = appTwo.GET(key)
			Expect(got).To(Equal(value))

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("checking it still works")
			key = helpers.RandomHex()
			value = helpers.RandomHex()
			appOne.PUT(value, key)
			got = appTwo.GET(key)
			Expect(got).To(Equal(value))
		})
	})
})