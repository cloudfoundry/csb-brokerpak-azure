package upgrade_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group DB Subsume", Label("mssql-db-failover-group-subsume"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance using the MASB broker")
		masbDBName := random.Name(random.WithPrefix("db"))
		masbServiceInstance := services.CreateInstance(
			"azure-sqldb",
			"StandardS0",
			services.WithMASBBroker(),
			services.WithParameters(map[string]string{
				"sqlServerName": metadata.PreProvisionedSQLServer,
				"sqldbName":     masbDBName,
				"resourceGroup": metadata.ResourceGroup,
			}),
		)
		defer masbServiceInstance.Delete()

		By("creating a failover group using the MASB broker")
		fogName := random.Name(random.WithPrefix("fog"))
		masbFOGServiceInstance := services.CreateInstance(
			"azure-sqldb-failover-group",
			"SecondaryDatabaseWithFailoverGroup",
			services.WithMASBBroker(),
			services.WithParameters(masbFOGConfig(masbDBName, fogName)),
		)
		defer masbFOGServiceInstance.Delete()

		By("pushing the unstarted app")
		app := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(app)

		By("binding the app to the MASB fog")
		masbinding := masbFOGServiceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating a schema using the app")
		schema := random.Name(random.WithMaxLength(10))
		app.PUT("", schema)

		By("setting a key-value using the app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		app.PUT(value, "%s/%s", schema, key)

		By("getting the value using the app")
		got := app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("Create the CSB with DB server details")
		serverPairTag := random.Name(random.WithMaxLength(10))
		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-fog-db"),
			brokers.WithSourceDir(releasedBuildDir),
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serverPairsConfig(serverPairTag)}),
		)
		defer serviceBroker.Delete()

		By("subsuming the database failover group")
		dbFogInstance := services.CreateInstance(
			"csb-azure-mssql-db-failover-group",
			"subsume",
			services.WithBroker(serviceBroker),
			services.WithParameters(map[string]string{
				"azure_primary_db_id":   fetchResourceID("db", masbDBName, metadata.PreProvisionedSQLServer),
				"azure_secondary_db_id": fetchResourceID("db", masbDBName, metadata.PreProvisionedFOGServer),
				"azure_fog_id":          fetchResourceID("failover-group", fogName, metadata.PreProvisionedSQLServer),
				"server_pair":           serverPairTag,
			}),
		)
		defer dbFogInstance.Delete()

		By("purging the MASB FOG instance")
		masbinding.Unbind()
		masbFOGServiceInstance.Purge()

		By("updating to another plan")
		dbFogInstance.Update("-p", "small")

		By("pushing the development version of the broker")
		serviceBroker.UpgradeBroker(developmentBuildDir)

		By("upgrading service instance")
		dbFogInstance.Upgrade()

		By("updating the instance plan after upgrading the service instance")
		dbFogInstance.Update("-p", "medium")

		By("binding the app to the CSB service instance")
		binding := dbFogInstance.Bind(app)
		defer apps.Delete(app) // app needs to be deleted before service instance

		By("restaging the app")
		apps.Restage(app)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("getting the value set with the MASB binding")
		got = app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the app")
		app.DELETE(schema)
	})
})

func masbFOGConfig(masbDBName, fogName string) any {
	return map[string]any{
		"primaryServerName":   metadata.PreProvisionedSQLServer,
		"primaryDbName":       masbDBName,
		"secondaryServerName": metadata.PreProvisionedFOGServer,
		"failoverGroupName":   fogName,
		"readWriteEndpoint": map[string]any{
			"failoverPolicy":                         "Automatic",
			"failoverWithDataLossGracePeriodMinutes": 60,
		},
	}
}

func serverPairsConfig(serverPairTag string) any {
	return map[string]any{
		serverPairTag: map[string]any{
			"admin_username": metadata.PreProvisionedSQLUsername,
			"admin_password": metadata.PreProvisionedSQLPassword,
			"primary": map[string]string{
				"server_name":    metadata.PreProvisionedSQLServer,
				"resource_group": metadata.ResourceGroup,
			},
			"secondary": map[string]string{
				"server_name":    metadata.PreProvisionedFOGServer,
				"resource_group": metadata.ResourceGroup,
			},
		},
	}
}
