package acceptance_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Server and DB", Label("mssql-server"), func() {
	It("can be accessed by an app", func() {
		serverConfig := newDatabaseServer()

		By("Create CSB with server details")
		serverTag := random.Name(random.WithMaxLength(10))
		creds := serverConfig.getMASBServerDetails(serverTag)

		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-db"),
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: creds}),
		)
		defer serviceBroker.Delete()

		By("creating a server")
		serverInstance := services.CreateInstance(
			"csb-azure-mssql-server",
			"standard",
			services.WithBroker(serviceBroker),
			services.WithParameters(serverConfig),
		)
		defer serverInstance.Delete()

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

func (d databaseServer) getMASBServerDetails(tag string) map[string]any {
	return map[string]any{
		tag: map[string]string{
			"server_name":           d.Name,
			"server_resource_group": metadata.ResourceGroup,
			"admin_username":        d.Username,
			"admin_password":        d.Password,
		},
	}
}
