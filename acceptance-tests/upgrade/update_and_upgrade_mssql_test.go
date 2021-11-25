package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlTest", func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := helpers.CreateBroker(
				helpers.BrokerWithPrefix("csb-mssql"),
				helpers.BrokerFromDir(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := helpers.CreateServiceFromBroker("csb-azure-mssql", "small-v2", serviceBroker.Name)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.MSSQL)
			appTwo := helpers.AppPushUnstarted(apps.MSSQL)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding to the apps")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)

			By("starting the apps")
			helpers.AppStart(appOne, appTwo)

			By("creating a schema using the first app")
			schema := helpers.RandomShortName()
			appOne.PUT("", schema)

			By("setting a key-value using the first app")
			keyOne := helpers.RandomHex()
			valueOne := helpers.RandomHex()
			appOne.PUT(valueOne, "%s/%s", schema, keyOne)

			By("getting the value using the second app")
			got := appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("checking previously written data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("creating a schema using the first app")
			schema = helpers.RandomShortName()
			appOne.PUT("", schema)

			By("checking data can still be written and read")
			keyThree := helpers.RandomHex()
			valueThree := helpers.RandomHex()
			appOne.PUT(valueThree, "%s/%s", schema, keyThree)

			got = appTwo.GET("%s/%s", schema, keyThree)
			Expect(got).To(Equal(valueThree))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)
		})
	})
})
