package upgrade_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMysqlTest", Label("mysql"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-mysql"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-azure-mysql",
				"small",
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MySQL))
			appTwo := apps.Push(apps.WithApp(apps.MySQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)
			apps.Start(appOne, appTwo)

			By("setting a key-value using the first app")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
			appOne.PUT(valueOne, keyOne)

			By("getting the value using the second app")
			got := appTwo.GET(keyOne)
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("checking data inserted before broker upgrade")
			Expect(appTwo.GET(keyOne)).To(Equal(valueOne))

			By("updating the instance plan")
			serviceInstance.Update("-p", "medium")

			By("checking data inserted before broker upgrade")
			Expect(appTwo.GET(keyOne)).To(Equal(valueOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("creating new data - post upgrade")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, keyTwo)
			got = appTwo.GET(keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("checking data inserted before broker upgrade")
			Expect(appTwo.GET(keyOne)).To(Equal(valueOne))

			By("updating the instance plan")
			serviceInstance.Update("-p", "medium")

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
