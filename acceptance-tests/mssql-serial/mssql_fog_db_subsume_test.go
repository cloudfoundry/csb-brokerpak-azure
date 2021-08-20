package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group DB Subsume", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance using the MASB broker")
		masbDBName := helpers.RandomName("db")
		masbDBInstance := helpers.CreateService("azure-sqldb", "StandardS0", masbServerConfig(masbDBName))
		defer masbDBInstance.Delete()

		By("creating a failover group using the MASB broker")
		fogName := helpers.RandomName("fog")
		masbFOGInstance := helpers.CreateService("azure-sqldb-failover-group", "SecondaryDatabaseWithFailoverGroup", masbFOGConfig(masbDBName, fogName))
		defer masbFOGInstance.Delete()

		By("pushing the unstarted app twice")
		app := helpers.AppPushUnstarted(apps.MSSQL)
		defer helpers.AppDelete(app)

		By("binding the app to the MASB service instance")
		masbDBInstance.Bind(app)

		By("starting the apps")
		helpers.AppStart(app)

		By("creating a schema using the app")
		schema := helpers.RandomShortName()
		app.PUT("", schema)

		By("setting a key-value using the app")
		key := helpers.RandomHex()
		value := helpers.RandomHex()
		app.PUT(value, "%s/%s", schema, key)

		By("getting the value using the app")
		got := app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("reconfiguring the CSB with DB server details")
		serverPairTag := helpers.RandomShortName()
		reconfigureCSBWithServerDetails(serverPairTag)

		By("subsuming the database failover group")
		dbFogInstance := helpers.CreateService("csb-azure-mssql-db-failover-group", "subsume", map[string]string{
			"azure_primary_db_id":   fetchResourceID("db", masbDBName, metadata.PreProvisionedSQLServer),
			"azure_secondary_db_id": fetchResourceID("db", masbDBName, metadata.PreProvisionedFOGServer),
			"azure_fog_id":          fetchResourceID("failover-group", fogName, metadata.PreProvisionedSQLServer),
			"server_pair":           serverPairTag,
		})
		defer dbFogInstance.Delete()

		By("purging the MASB service instance")
		helpers.CF("purge-service-instance", "-f", masbFOGInstance.Name())

		By("updating to another plan")
		dbFogInstance.UpdateService("-p", "small")

		By("binding the app to the CSB service instance")
		binding := dbFogInstance.Bind(app)
		defer helpers.AppDelete(app) // app needs to be deleted before service instance

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

		By("getting the value set with the MASB binding")
		got = app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the app")
		app.DELETE(schema)
	})
})

func masbFOGConfig(masbDBName, fogName string) interface{} {
	return map[string]interface{}{
		"primaryServerName":   metadata.PreProvisionedSQLServer,
		"primaryDbName":       masbDBName,
		"secondaryServerName": metadata.PreProvisionedFOGServer,
		"failoverGroupName":   fogName,
		"readWriteEndpoint": map[string]interface{}{
			"failoverPolicy":                         "Automatic",
			"failoverWithDataLossGracePeriodMinutes": 60,
		},
	}
}

func serverPairsConfig(serverPairTag string) interface{} {
	return map[string]interface{}{
		serverPairTag: map[string]interface{}{
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

func reconfigureCSBWithServerDetails(serverPairTag string) {
	helpers.SetBrokerEnv(
		helpers.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serverPairsConfig(serverPairTag)},
		helpers.EnvVar{Name: "GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS", Value: map[string]interface{}{"server_credential_pairs": serverPairsConfig(serverPairTag)}},
	)
}
