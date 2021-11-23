package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlDBTest", func() {
	Context("When upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			brokerName := helpers.RandomName("csb-mssql-db")
			serviceBroker := helpers.PushAndStartBroker(brokerName, releasedBuildDir)
			defer serviceBroker.Delete()

			By("creating a service")
			serverConfig := newDatabaseServer()
			serverInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", brokerName, serverConfig)
			defer serverInstance.Delete()

			By("reconfiguring the CSB with DB server details")
			serverTag := serverConfig.reconfigureCSBWithServerDetails(brokerName)

			By("creating a database in the server")
			dbInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db", "small", brokerName, map[string]string{"server": serverTag})
			defer dbInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.MSSQL)
			appTwo := helpers.AppPushUnstarted(apps.MSSQL)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding to the apps")
			dbInstance.Bind(appOne)
			dbInstance.Bind(appTwo)

			By("starting the apps")
			helpers.AppStart(appOne, appTwo)

			By("creating a schema using the first app")
			schema := helpers.RandomShortName()
			appOne.PUT("", schema)

			By("setting a key-value using the first app")
			keyOne := helpers.RandomHex()
			valueOne := helpers.RandomHex()
			appOne.PUT(valueOne, "%s/%s", schema, keyOne)

			By("getting the value using the second app")
			got := appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("updating the instance plan")
			dbInstance.UpdateService("-p", "medium")

			By("checking previously created data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)

			By("deleting bindings created before the upgrade")
			dbInstance.Unbind(appOne)
			dbInstance.Unbind(appTwo)

			By("creating new bindings")
			dbInstance.Bind(appOne)
			dbInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("creating a schema using the first app")
			schema = helpers.RandomShortName()
			appOne.PUT("", schema)

			By("checking data can still be written and read")
			keyTwo := helpers.RandomHex()
			valueTwo := helpers.RandomHex()
			appOne.PUT(valueTwo, "%s/%s", schema, keyTwo)

			got = appTwo.GET("%s/%s", schema, keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)
		})
	})
})

func newDatabaseServer() databaseServer {
	return databaseServer{
		Name:     helpers.RandomName("server"),
		Username: helpers.RandomShortName(),
		Password: helpers.RandomPassword(),
	}
}

type databaseServer struct {
	Name     string `json:"instance_name"`
	Username string `json:"admin_username"`
	Password string `json:"admin_password"`
}

func (d databaseServer) reconfigureCSBWithServerDetails(broker string) string {
	serverTag := helpers.RandomShortName()

	creds := map[string]interface{}{
		serverTag: map[string]string{
			"server_name":           d.Name,
			"server_resource_group": metadata.ResourceGroup,
			"admin_username":        d.Username,
			"admin_password":        d.Password,
		},
	}

	helpers.SetBrokerEnv(broker, helpers.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: creds})

	helpers.RestartBroker(broker)

	return serverTag
}

var metadata struct {
	ResourceGroup             string `jsonry:"name"`
	PreProvisionedSQLUsername string `jsonry:"masb_config.pre_provisioned_sql.username"`
	PreProvisionedSQLPassword string `jsonry:"masb_config.pre_provisioned_sql.password"`
	PreProvisionedSQLServer   string `jsonry:"masb_config.pre_provisioned_sql.server_name"`
	PreProvisionedSQLLocation string `jsonry:"masb_config.location"`
	PreProvisionedFOGUsername string `jsonry:"masb_config.pre_provisioned_fog_sql.username"`
	PreProvisionedFOGPassword string `jsonry:"masb_config.pre_provisioned_fog_sql.password"`
	PreProvisionedFOGServer   string `jsonry:"masb_config.pre_provisioned_fog_sql.server_name"`
	PreProvisionedFOGLocation string `jsonry:"masb_config.pre_provisioned_fog_sql.location"`
}
