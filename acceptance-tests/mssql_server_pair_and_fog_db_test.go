package acceptance_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/serverpairs"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Server Pair and Failover Group DB", Label("mssql-server-pair"), func() {
	It("can be accessed by an app", func() {
		serversConfig := serverpairs.NewDatabaseServerPair(metadata)

		By("Create CSB with server details")
		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-db"),
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
		)
		defer serviceBroker.Delete()

		By("creating a primary server")
		serverInstancePrimary := services.CreateInstance(
			"csb-azure-mssql-server",
			"standard",
			services.WithBroker(serviceBroker),
			services.WithParameters(serversConfig.PrimaryConfig()),
		)
		defer serverInstancePrimary.Delete()

		// We have previously experienced problems with the CF CLI when doing things in parallel
		By("creating a secondary server in a different resource group")
		secondaryResourceGroupInstance := services.CreateInstance(
			"csb-azure-resource-group",
			"standard",
			services.WithBroker(serviceBroker),
			services.WithParameters(serversConfig.SecondaryResourceGroupConfig()),
		)
		defer secondaryResourceGroupInstance.Delete()

		serverInstanceSecondary := services.CreateInstance(
			"csb-azure-mssql-server",
			"standard",
			services.WithBroker(serviceBroker),
			services.WithParameters(serversConfig.SecondaryConfig()),
		)
		defer serverInstanceSecondary.Delete()

		By("creating a database failover group on the server pair")
		fogName := random.Name(random.WithPrefix("fog"))
		dbFogInstance := services.CreateInstance(
			"csb-azure-mssql-db-failover-group",
			"small",
			services.WithBroker(serviceBroker),
			services.WithParameters(map[string]string{
				"server_pair":   serversConfig.ServerPairTag,
				"instance_name": fogName,
			}),
		)
		defer dbFogInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.MSSQL))
		appTwo := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := dbFogInstance.Bind(appOne)
		dbFogInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

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

		By("triggering failover")
		failoverServiceInstance := services.CreateInstance(
			"csb-azure-mssql-fog-run-failover",
			"standard",
			services.WithBroker(serviceBroker),
			services.WithParameters(map[string]interface{}{
				"server_pair_name":  serversConfig.ServerPairTag,
				"server_pairs":      serversConfig.ServerPairsConfig(),
				"fog_instance_name": fogName,
			}),
		)
		defer failoverServiceInstance.Delete()

		By("setting another key-value")
		keyTwo := random.Hexadecimal()
		valueTwo := random.Hexadecimal()
		appTwo.PUT(valueTwo, "%s/%s", schema, keyTwo)

		By("getting the previously set values")
		Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))
		Expect(appTwo.GET("%s/%s", schema, keyTwo)).To(Equal(valueTwo))

		By("reverting the failover")
		failoverServiceInstance.Delete()

		By("dropping the schema")
		appOne.DELETE(schema)
	})
})
