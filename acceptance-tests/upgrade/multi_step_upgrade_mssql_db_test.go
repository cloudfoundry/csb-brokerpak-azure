package upgrade_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"
)

var _ = Describe("MultiStepUpgradeMssqlDBTest", Label("multi-step"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-srvdb"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serverConfig := newDatabaseServer()
			serverInstance := services.CreateInstance(
				"csb-azure-mssql-server",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(serverConfig),
			)
			defer serverInstance.Delete()

			By("reconfiguring the CSB with DB server details")
			serverTag := random.Name(random.WithMaxLength(10))
			serviceBroker.UpdateEnv(apps.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: serverConfig.serverDetails(serverTag)})

			By("creating a database in the server")
			dbInstance := services.CreateInstance(
				"csb-azure-mssql-db",
				"small",
				services.WithBroker(serviceBroker),
				services.WithParameters(map[string]string{"server": serverTag}),
			)
			defer dbInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MSSQL))
			appTwo := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := dbInstance.Bind(appOne)
			bindingTwo := dbInstance.Bind(appTwo)

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

			By("Performing all intermediate upgrades")
			for _, brokerDir := range strings.Split(intermediateBuildDirs, ",") {
				By("pushing the next version of the broker")
				serviceBroker.UpgradeBroker(brokerDir)
				By("upgrading service instance")
				dbInstance.Upgrade()
				serverInstance.Upgrade()
			}

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)
			By("upgrading service instance")
			dbInstance.Upgrade()
			serverInstance.Upgrade()

			By("checking previously created data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("updating the instance plan")
			dbInstance.Update("-p", "medium")

			By("checking previously created data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings")
			dbInstance.Bind(appOne)
			dbInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("creating a schema using the first app")
			schema = random.Name(random.WithMaxLength(10))
			appOne.PUT("", schema)

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, "%s/%s", schema, keyTwo)

			got = appTwo.GET("%s/%s", schema, keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)
		})
	})
})
