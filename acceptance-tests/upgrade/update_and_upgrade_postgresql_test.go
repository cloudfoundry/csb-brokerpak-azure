package upgrade_test

import (
	"acceptancetests/helpers"
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradePostgreSQLTest", func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := helpers.CreateBroker(
				helpers.BrokerWithPrefix("csb-postgresql"),
				helpers.BrokerFromDir(releasedBuildDir),
				helpers.BrokerWithEnv(apps.EnvVar{Name: "BROKERPAK_UPDATES_ENABLED", Value: true}),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := helpers.CreateServiceFromBroker("csb-azure-postgresql", "small", serviceBroker.Name)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.PostgreSQL))
			appTwo := apps.Push(apps.WithApp(apps.PostgreSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

			By("creating a schema using the first app")
			schema := random.Name(random.WithMaxLength(10))
			appOne.PUT("", schema)

			By("setting a key-value using the first app")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
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

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking previously written data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("creating a schema using the first app")
			schemaTwo := random.Name(random.WithMaxLength(10))
			appOne.PUT("", schemaTwo)

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, "%s/%s", schemaTwo, keyTwo)

			got = appTwo.GET("%s/%s", schemaTwo, keyTwo)
			Expect(got).To(Equal(valueTwo))
		})
	})
})
