package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMysqlTest", func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := helpers.CreateBroker(
				helpers.BrokerWithPrefix("csb-mysql"),
				helpers.BrokerFromDir(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := helpers.CreateServiceFromBroker("csb-azure-mysql", "small", serviceBroker.Name)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.MySQL)
			appTwo := helpers.AppPushUnstarted(apps.MySQL)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding to the apps")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppStart(appOne, appTwo)

			By("setting a key-value using the first app")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
			appOne.PUT(valueOne, keyOne)

			By("getting the value using the second app")
			got := appTwo.GET(keyOne)
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("creating new data - post upgrade")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, keyTwo)
			got = appTwo.GET(keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("checking data inserted before broker upgrade")
			Expect(appTwo.GET(keyOne)).To(Equal(valueOne))

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("checking previously written data is still accessible")
			got = appTwo.GET(keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("checking data can still be written and read")
			keyThree := random.Hexadecimal()
			valueThree := random.Hexadecimal()
			appOne.PUT(valueThree, keyThree)
			got = appTwo.GET(keyThree)
			Expect(got).To(Equal(valueThree))
		})
	})
})
