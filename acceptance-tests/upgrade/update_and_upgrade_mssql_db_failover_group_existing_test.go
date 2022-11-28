package upgrade_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upgrade and Update csb-azure-mssql-db-failover-group 'existing' plan", Label("mssql-db-failover-group-existing"), func() {
	When("upgrading broker version", Label("modern"), func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			rgConfig := resourceGroupConfig()
			serversConfig := newServerPair(rgConfig.Name)

			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-db-fo"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
			)
			defer serviceBroker.Delete()

			By("creating a new resource group")
			resourceGroupInstance := services.CreateInstance(
				"csb-azure-resource-group",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(rgConfig),
			)
			defer resourceGroupInstance.Delete()

			By("creating primary and secondary DB servers in the resource group")
			serverInstancePrimary := services.CreateInstance(
				"csb-azure-mssql-server",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(serversConfig.PrimaryConfig()),
			)
			defer serverInstancePrimary.Delete()

			serverInstanceSecondary := services.CreateInstance(
				"csb-azure-mssql-server",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(serversConfig.SecondaryConfig()),
			)
			defer serverInstanceSecondary.Delete()

			By("creating a failover group service instance")
			fogConfig := map[string]any{
				"instance_name": random.Name(random.WithPrefix("fog")),
				"db_name":       random.Name(random.WithPrefix("db")),
				"server_pair":   serversConfig.ServerPairTag,
			}

			initialFogInstance := services.CreateInstance(
				"csb-azure-mssql-db-failover-group",
				"medium",
				services.WithBroker(serviceBroker),
				services.WithParameters(fogConfig),
			)
			defer initialFogInstance.Delete()

			By("creating a failover group service instance with 'existing' plan")
			existingFogInstance := services.CreateInstance(
				"csb-azure-mssql-db-failover-group",
				"existing",
				services.WithBroker(serviceBroker),
				services.WithParameters(fogConfig),
			)
			defer existingFogInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MSSQL))
			appTwo := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			existingFogInstance.Bind(appOne)
			existingFogInstance.Bind(appTwo)

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

			By("upgrading previous services")
			resourceGroupInstance.Upgrade()
			serverInstancePrimary.Upgrade()
			serverInstanceSecondary.Upgrade()
			initialFogInstance.Upgrade()
			existingFogInstance.Upgrade()

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("updating the instance plan and adding the 'existing' property")
			existingFogInstance.Update("-p", "medium", "-c", `{"existing": true}`)

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))
		})
	})
})
