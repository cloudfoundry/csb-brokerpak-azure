package acceptance_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/serverpairs"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group Existing", Label("mssql-failover-group"), func() {
	It("can be accessed by an app", func() {
		By("deploying the CSB")
		serversConfig := serverpairs.NewDatabaseServerPair(metadata)
		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-db-fog"),
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
		)
		defer serviceBroker.Delete()

		By("creating a new resource group")
		resourceGroupInstance := services.CreateInstance(
			"csb-azure-resource-group",
			"standard",
			services.WithBroker(serviceBroker),
			services.WithParameters(serversConfig.SecondaryResourceGroupConfig()),
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
		fogConfig := map[string]string{
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

		By("pushing an unstarted app")
		app := apps.Push(apps.WithApp(apps.MSSQL))

		By("binding the app to the initial failover group service instance")
		initialFogInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating a schema")
		schema := random.Name(random.WithMaxLength(10))
		app.PUT("", "%s?dbo=false", schema)

		By("setting a key-value")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		app.PUT(value, "%s/%s", schema, key)

		By("connecting to the existing failover group")
		dbFogInstance := services.CreateInstance(
			"csb-azure-mssql-db-failover-group",
			"existing",
			services.WithBroker(serviceBroker),
			services.WithParameters(fogConfig),
		)
		defer dbFogInstance.Delete()

		By("purging the initial FOG instance")
		initialFogInstance.Purge()

		By("binding the app to the CSB service instance")
		bindingTwo := dbFogInstance.Bind(app)
		defer apps.Delete(app) // app needs to be deleted before service instance

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingTwo.Credential()).To(matchers.HaveCredHubRef)

		By("getting the value set with the initial binding")
		got := app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the app")
		app.DELETE(schema)
	})
})
