package upgrade_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlDBTest", Label("mssql-db"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-srvdb"),
				brokers.WithSourceDir(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serverConfig := newDatabaseServer()
			serverInstance := services.CreateInstance(
				"csb-azure-mssql-server",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(serverConfig),
			)
			defer serverInstance.Delete()

			By("reconfiguring the CSB with DB server details")
			serverTag := serverConfig.reconfigureCSBWithServerDetails(serviceBroker)

			By("creating a database in the server")
			dbInstance := services.CreateInstance(
				"csb-azure-mssql-db",
				"small",
				services.WithBroker(serviceBroker),
				services.WithParameters(map[string]string{"server": serverTag}),
			)
			defer dbInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MSSQL))
			appTwo := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := dbInstance.Bind(appOne)
			bindingTwo := dbInstance.Bind(appTwo)

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
			serviceBroker.UpdateSourceDir(developmentBuildDir)

			By("updating the instance plan")
			dbInstance.Update("-p", "medium")

			By("checking previously created data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings")
			dbInstance.Bind(appOne)
			dbInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("creating a schema using the first app")
			schema = random.Name(random.WithMaxLength(10))
			appOne.PUT("", schema)

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
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
		Name:     random.Name(random.WithPrefix("server")),
		Username: random.Name(random.WithMaxLength(10)),
		Password: random.Password(),
	}
}

type databaseServer struct {
	Name     string `json:"instance_name"`
	Username string `json:"admin_username"`
	Password string `json:"admin_password"`
}

func (d databaseServer) reconfigureCSBWithServerDetails(broker *brokers.Broker) string {
	serverTag := random.Name(random.WithMaxLength(10))

	creds := map[string]interface{}{
		serverTag: map[string]string{
			"server_name":           d.Name,
			"server_resource_group": metadata.ResourceGroup,
			"admin_username":        d.Username,
			"admin_password":        d.Password,
		},
	}

	broker.UpdateEnv(apps.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: creds})

	return serverTag
}
