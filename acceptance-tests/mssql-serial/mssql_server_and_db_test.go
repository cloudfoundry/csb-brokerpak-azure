package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Server and DB", func() {
	It("can be accessed by an app", func() {
		By("creating a server")
		serverConfig := newDatabaseServer()
		serverInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", helpers.DefaultBroker().Name, serverConfig)
		defer serverInstance.Delete()

		By("reconfiguring the CSB with DB server details")
		serverTag := serverConfig.reconfigureCSBWithServerDetails()

		By("creating a database in the server")
		dbInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db", "small", helpers.DefaultBroker().Name, map[string]string{"server": serverTag})
		defer dbInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.MSSQL)
		appTwo := helpers.AppPushUnstarted(apps.MSSQL)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := dbInstance.Bind(appOne)
		dbInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

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

func (d databaseServer) reconfigureCSBWithServerDetails() string {
	serverTag := random.Name(random.WithMaxLength(10))

	creds := map[string]interface{}{
		serverTag: map[string]string{
			"server_name":           d.Name,
			"server_resource_group": metadata.ResourceGroup,
			"admin_username":        d.Username,
			"admin_password":        d.Password,
		},
	}

	helpers.SetBrokerEnvAndRestart(
		helpers.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: creds},
	)

	return serverTag
}
