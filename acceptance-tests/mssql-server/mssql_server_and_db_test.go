package mssql_server_test

import (
	"acceptancetests/helpers"
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Server and DB", func() {
	It("can be accessed by an app", func() {
		serverConfig := newDatabaseServer()

		By("Create CSB with server details")
		serverTag := random.Name(random.WithMaxLength(10))
		creds := serverConfig.getMASBServerDetails(serverTag)

		serviceBroker := helpers.CreateBroker(
			helpers.BrokerWithPrefix("csb-mssql-db"),
			helpers.BrokerWithEnv(helpers.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: creds}),
		)
		defer serviceBroker.Delete()

		By("creating a server")
		serverInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", serviceBroker.Name, serverConfig)
		defer serverInstance.Delete()

		By("creating a database in the server")
		dbInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db", "small", serviceBroker.Name, map[string]string{"server": serverTag})
		defer dbInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.MSSQL))
		appTwo := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := dbInstance.Bind(appOne)
		dbInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("creating a schema using the first app")
		schema := random.Name(random.WithMaxLength(10))
		appOne.PUT("", schema)

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appOne.PUT(value, "%s/%s", schema, key)

		By("getting the value using the second app")
		got := appTwo.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the first app")
		appOne.DELETE(schema)
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

func (d databaseServer) getMASBServerDetails(tag string) map[string]interface{} {
	return map[string]interface{}{
		tag: map[string]string{
			"server_name":           d.Name,
			"server_resource_group": metadata.ResourceGroup,
			"admin_username":        d.Username,
			"admin_password":        d.Password,
		},
	}
}
