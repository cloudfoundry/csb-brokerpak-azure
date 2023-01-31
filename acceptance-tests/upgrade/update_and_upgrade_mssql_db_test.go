package upgrade_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlDBTest", Label("mssql-db"), func() {
	When("upgrading broker version", Label("modern"), func() {
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
			serverTag := random.Name(random.WithMaxLength(10))
			serviceBroker.UpdateEnv(apps.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: serverConfig.serverDetails(serverTag)})

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
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading service instance")
			dbInstance.Upgrade()
			serverInstance.Upgrade()

			By("checking previously created data still accessible")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

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
	When("using a config file for broker configuration", func() {
		It("it should respect the config as if it were set via an env var", func() {
			By("pushing a broker with a config file")

			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-srvdb"),
				brokers.WithSourceDir(developmentBuildDir),
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

			By("reconfiguring the CSB with DB server details and pushing it")
			serverTag := random.Name(random.WithMaxLength(10))
			serverCreds := serverConfig.serverDetails(serverTag)
			serviceBroker.UpdateConfig(map[string]interface{}{
				"azure.mssql_db_server_creds": serverCreds,
			})

			By("creating a database in the server")
			dbInstance := services.CreateInstance(
				"csb-azure-mssql-db",
				"small",
				services.WithBroker(serviceBroker),
				services.WithParameters(map[string]any{
					"server": serverTag,
				}),
			)
			defer dbInstance.Delete()

			By("pushing the unstarted app")
			app := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(app)

			By("binding to the app")
			binding := dbInstance.Bind(app)

			By("starting the app")
			apps.Start(app)

			By("creating a schema")
			schema := random.Name(random.WithMaxLength(10))
			app.PUT("", "%s?dbo=false", schema)

			By("setting a key-value")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
			app.PUT(valueOne, "%s/%s", schema, keyOne)

			By("getting the value")
			got := app.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("deleting the schema so we can unbind")
			app.DELETE(schema)

			By("deleting bindings")
			binding.Unbind()

		})

	})
	When("upgrading broker version", Label("ancient"), func() {
		It("should continue to work", func() {
			By("pushing an ancient broker version")
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
			serverTag := random.Name(random.WithMaxLength(10))
			serverCreds := serverConfig.serverDetails(serverTag)
			serviceBroker.UpdateEnv(
				apps.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: serverCreds},
				apps.EnvVar{Name: "GSB_SERVICE_CSB_AZURE_MSSQL_DB_PROVISION_DEFAULTS", Value: map[string]any{"server_credentials": serverCreds}},
			)

			By("creating a database in the server")
			dbInstance := services.CreateInstance(
				"csb-azure-mssql-db",
				"small",
				services.WithBroker(serviceBroker),
				services.WithParameters(map[string]any{
					"server":             serverTag,
					"server_credentials": serverCreds,
				}),
			)
			defer dbInstance.Delete()

			By("pushing the unstarted app")
			app := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(app)

			By("binding to the app")
			binding := dbInstance.Bind(app)

			By("starting the app")
			apps.Start(app)

			By("creating a schema")
			schema := random.Name(random.WithMaxLength(10))
			app.PUT("", "%s?dbo=false", schema)

			By("setting a key-value")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
			app.PUT(valueOne, "%s/%s", schema, keyOne)

			By("getting the value")
			got := app.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading service instance")
			dbInstance.Upgrade()
			serverInstance.Upgrade()

			By("checking previously created data still accessible")
			got = app.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("updating the instance plan")
			dbInstance.Update("-p", "medium")

			By("checking previously created data still accessible")
			got = app.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("dropping the schema used to allow us to unbind")
			app.DELETE(schema)

			By("deleting bindings created before the upgrade")
			binding.Unbind()

			By("creating new bindings")
			dbInstance.Bind(app)
			apps.Restage(app)

			By("creating a schema")
			schema = random.Name(random.WithMaxLength(10))
			app.PUT("", schema)

			By("checking data can still be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			app.PUT(valueTwo, "%s/%s", schema, keyTwo)

			got = app.GET("%s/%s", schema, keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("dropping the schema used to allow us to unbind")
			app.DELETE(schema)
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

func (d databaseServer) serverDetails(serverTag string) map[string]any {
	creds := map[string]any{
		serverTag: map[string]string{
			"server_name":           d.Name,
			"server_resource_group": metadata.ResourceGroup,
			"admin_username":        d.Username,
			"admin_password":        d.Password,
		},
	}

	return creds
}
