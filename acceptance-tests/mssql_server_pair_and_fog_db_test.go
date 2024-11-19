package acceptance_test

import (
	"context"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/mssqlserver"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Server Pair and Failover Group DB", Label("mssql-db-failover-group"), func() {
	It("can be accessed by an app", func() {
		ctx := context.Background()

		By("creating primary and secondary DB servers in their resource group")
		serversConfig, err := mssqlserver.CreateServerPair(ctx, metadata, subscriptionID)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			By("deleting the created resource group and DB servers")
			_ = mssqlserver.Cleanup(ctx, serversConfig, subscriptionID)
		})

		By("Create CSB with server details")
		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-db"),
			brokers.WithLatestEnv(),
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
		)
		defer serviceBroker.Delete()

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
		appOne.PUT("", "%s?dbo=false", schema)

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
			services.WithParameters(map[string]any{
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

		By("deleting binding one the binding two keeps reading the value - object reassignment works")
		binding.Unbind()
		Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

		By("reverting the failover")
		failoverServiceInstance.Delete()

		By("dropping the schema using the second app")
		appTwo.DELETE(schema)
	})
})
