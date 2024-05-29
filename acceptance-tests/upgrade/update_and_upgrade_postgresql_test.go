package upgrade_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradePostgreSQLTest", Label("postgresql"), func() {
	When("upgrading broker version", func() {
		Context("schema supported", Label("schema"), func() {
			It("should continue to work", func() {
				By("pushing latest released broker version")
				serviceBroker := brokers.Create(
					brokers.WithPrefix("csb-postgresql"),
					brokers.WithSourceDir(releasedBuildDir),
					brokers.WithReleaseEnv(releasedBuildDir),
				)
				defer serviceBroker.Delete()

				By("creating a service")
				serviceInstance := services.CreateInstance(
					"csb-azure-postgresql",
					"small",
					services.WithBroker(serviceBroker),
				)
				defer serviceInstance.Delete()

				By("pushing the unstarted app twice")
				appOne := apps.Push(apps.WithApp(apps.PostgreSQL))
				appTwo := apps.Push(apps.WithApp(apps.PostgreSQL))
				defer apps.Delete(appOne, appTwo)

				By("binding to the apps")
				bindingOne := serviceInstance.Bind(appOne)
				bindingTwo := serviceInstance.Bind(appTwo)

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
				serviceBroker.UpgradeBroker(developmentBuildDir)

				By("upgrading service instance")
				serviceInstance.Upgrade()

				By("checking previously written data still accessible")
				got = appTwo.GET("%s/%s", schema, keyOne)
				Expect(got).To(Equal(valueOne))

				By("deleting bindings created before the upgrade")
				bindingOne.Unbind()
				bindingTwo.Unbind()

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

		Context("schema not supported", Label("noschema"), func() {
			It("should continue to work", func() {
				By("pushing latest released broker version")
				serviceBroker := brokers.Create(
					brokers.WithPrefix("csb-postgresql"),
					brokers.WithSourceDir(releasedBuildDir),
				)
				defer serviceBroker.Delete()

				By("creating a service")
				serviceInstance := services.CreateInstance(
					"csb-azure-postgresql",
					"small",
					services.WithBroker(serviceBroker),
				)
				defer serviceInstance.Delete()

				By("pushing the unstarted app")
				app := apps.Push(apps.WithApp(apps.PostgreSQL))
				defer apps.Delete(app)

				By("binding to the app")
				binding := serviceInstance.Bind(app)

				By("starting the app")
				apps.Start(app)

				By("setting a key-value")
				keyOne := random.Hexadecimal()
				valueOne := random.Hexadecimal()
				const schema = "public"
				app.PUT(valueOne, "%s/%s", schema, keyOne)

				By("getting the value")
				got := app.GET("%s/%s", schema, keyOne)
				Expect(got).To(Equal(valueOne))

				By("pushing the development version of the broker")
				serviceBroker.UpgradeBroker(developmentBuildDir)

				By("upgrading service instance")
				serviceInstance.Upgrade()

				By("checking previously written data still accessible")
				got = app.GET("%s/%s", schema, keyOne)
				Expect(got).To(Equal(valueOne))

				By("deleting bindings created before the upgrade")
				binding.Unbind()

				By("creating new bindings and testing they still work")
				serviceInstance.Bind(app)
				apps.Restage(app)

				By("checking data can still be written and read")
				keyTwo := random.Hexadecimal()
				valueTwo := random.Hexadecimal()
				app.PUT(valueTwo, "%s/%s", schema, keyTwo)

				got = app.GET("%s/%s", schema, keyTwo)
				Expect(got).To(Equal(valueTwo))
			})
		})
	})
})
