package acceptance_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/cf"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL DB Subsume", Label("mssql-db", "subsume"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance using the MASB broker")
		masbDBName := random.Name(random.WithPrefix("db"))
		masbServiceInstance := services.CreateInstance(
			"azure-sqldb",
			"basic",
			services.WithMASBBroker(),
			services.WithParameters(map[string]string{
				"sqlServerName": metadata.PreProvisionedSQLServer,
				"sqldbName":     masbDBName,
				"resourceGroup": metadata.ResourceGroup,
			}),
		)
		defer masbServiceInstance.Delete()

		By("pushing the unstarted app")
		app := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(app)

		By("binding the app to the MASB service instance")
		masbServiceInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating a schema using the app")
		schema := random.Name(random.WithMaxLength(10))
		app.PUT("", schema)

		By("setting a key-value using the app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		app.PUT(value, "%s/%s", schema, key)

		By("fetching the Azure resource ID of the database")
		resource := fetchResourceID("db", masbDBName, metadata.PreProvisionedSQLServer)

		By("Create CSB with DB server details")
		serverTag := random.Name(random.WithMaxLength(10))
		creds := getMASBServerDetails(serverTag)

		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-db"),
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: creds}),
		)
		defer serviceBroker.Delete()

		By("subsuming the database")
		csbServiceInstance := services.CreateInstance(
			"csb-azure-mssql-db",
			"subsume",
			services.WithBroker(serviceBroker),
			services.WithParameters(subsumeDBParams(resource, serverTag)),
		)
		defer csbServiceInstance.Delete()

		By("purging the MASB service instance")
		cf.Run("purge-service-instance", "-f", masbServiceInstance.Name)

		By("updating to another plan")
		csbServiceInstance.Update("-p", "small")

		By("binding the app to the CSB service instance")
		binding := csbServiceInstance.Bind(app)
		defer apps.Delete(app) // app needs to be deleted before service instance

		By("restaging the app")
		apps.Restage(app)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("getting the value set with the MASB binding")
		got := app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the app")
		app.DELETE(schema)
	})
})

func subsumeDBParams(resource, serverTag string) any {
	return map[string]any{
		"azure_db_id": resource,
		"server":      serverTag,
	}
}

func getMASBServerDetails(tag string) map[string]any {
	return map[string]any{
		tag: map[string]string{
			"server_name":           metadata.PreProvisionedSQLServer,
			"server_resource_group": metadata.ResourceGroup,
			"admin_username":        metadata.PreProvisionedSQLUsername,
			"admin_password":        metadata.PreProvisionedSQLPassword,
		},
	}
}
