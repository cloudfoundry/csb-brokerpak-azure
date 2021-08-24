package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL DB Subsume", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance using the MASB broker")
		masbDBName := helpers.RandomName("db")
		masbServiceInstance := helpers.CreateService("azure-sqldb", "basic", masbServerConfig(masbDBName))
		defer masbServiceInstance.Delete()

		By("pushing the unstarted app")
		app := helpers.AppPushUnstarted(apps.MSSQL)
		defer helpers.AppDelete(app)

		By("binding the app to the MASB service instance")
		masbServiceInstance.Bind(app)

		By("starting the app")
		helpers.AppStart(app)

		By("creating a schema using the app")
		schema := helpers.RandomShortName()
		app.PUT("", schema)

		By("setting a key-value using the app")
		key := helpers.RandomHex()
		value := helpers.RandomHex()
		app.PUT(value, "%s/%s", schema, key)

		By("fetching the Azure resource ID of the database")
		resource := fetchResourceID("db", masbDBName, metadata.PreProvisionedSQLServer)

		By("reconfiguring the CSB with DB server details")
		serverTag := reconfigureCSBWithMASBServerDetails()

		By("subsuming the database")
		csbServiceInstance := helpers.CreateService("csb-azure-mssql-db", "subsume", subsumeDBParams(resource, serverTag))
		defer csbServiceInstance.Delete()

		By("purging the MASB service instance")
		helpers.CF("purge-service-instance", "-f", masbServiceInstance.Name())

		By("updating to another plan")
		csbServiceInstance.UpdateService("-p", "small")

		By("binding the app to the CSB service instance")
		binding := csbServiceInstance.Bind(app)
		defer helpers.AppDelete(app) // app needs to be deleted before service instance

		By("restaging the app")
		helpers.AppRestage(app)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

		By("getting the value set with the MASB binding")
		got := app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the app")
		app.DELETE(schema)
	})
})

func subsumeDBParams(resource, serverTag string) interface{} {
	return map[string]interface{}{
		"azure_db_id": resource,
		"server_name": metadata.PreProvisionedSQLServer,
		"server":      serverTag,
	}
}

func reconfigureCSBWithMASBServerDetails() string {
	tag := helpers.RandomShortName()
	creds := map[string]interface{}{
		tag: map[string]string{
			"server_name":           metadata.PreProvisionedSQLServer,
			"server_resource_group": metadata.ResourceGroup,
			"admin_username":        metadata.PreProvisionedSQLUsername,
			"admin_password":        metadata.PreProvisionedSQLPassword,
		},
	}

	helpers.SetBrokerEnv(
		helpers.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: creds},
		helpers.EnvVar{Name: "GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS", Value: map[string]interface{}{"server_credentials": creds}},
	)

	return tag
}

func masbServerConfig(dbName string) interface{} {
	return map[string]string{
		"sqlServerName": metadata.PreProvisionedSQLServer,
		"sqldbName":     dbName,
		"resourceGroup": metadata.ResourceGroup,
	}
}
