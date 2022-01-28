package mssql_db_failover_group_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"
	"os/exec"
	"strings"
	"time"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group DB Subsume", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance using the MASB broker")
		masbDBName := random.Name(random.WithPrefix("db"))
		masbDBInstance := services.CreateInstance(
			"azure-sqldb",
			"StandardS0",
			services.WithMASBBroker(),
			services.WithParameters(masbServerConfig(masbDBName)),
		)
		defer masbDBInstance.Delete()

		By("creating a failover group using the MASB broker")
		fogName := random.Name(random.WithPrefix("fog"))
		masbFOGInstance := services.CreateInstance(
			"azure-sqldb-failover-group",
			"SecondaryDatabaseWithFailoverGroup",
			services.WithMASBBroker(),
			services.WithParameters(masbFOGConfig(masbDBName, fogName)),
		)
		defer masbFOGInstance.Delete()

		By("pushing the unstarted app twice")
		app := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(app)

		By("binding the app to the MASB fog")
		masbFOGInstance.Bind(app)

		By("starting the apps")
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
			// Disable brokerpak_updates due to bug - https://www.pivotaltracker.com/story/show/180586187
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serverPairsConfig(serverPairTag)}, apps.EnvVar{Name: "BROKERPAK_UPDATES_ENABLED", Value: false}),
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
		cf.Run("purge-service-instance", "-f", masbFOGInstance.Name)

		By("updating to another plan")
		dbFogInstance.Update("-p", "small")

		By("binding the app to the CSB service instance")
		binding := dbFogInstance.Bind(app)
		defer apps.Delete(app) // app needs to be deleted before service instance

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

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

func masbServerConfig(dbName string) interface{} {
	return map[string]string{
		"sqlServerName": metadata.PreProvisionedSQLServer,
		"sqldbName":     dbName,
		"resourceGroup": metadata.ResourceGroup,
	}
}

func fetchResourceID(kind, name, server string) string {
	command := exec.Command("az", "sql", kind, "show", "--name", name, "--server", server, "--resource-group", metadata.ResourceGroup, "--query", "id", "-o", "tsv")
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, time.Minute).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}
