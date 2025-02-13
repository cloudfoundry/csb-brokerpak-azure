package acceptance_test

import (
	"context"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/mssqlserver"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Tests the *csb-azure-mssql-db* service offering
// Does NOT use the default broker: deploys a custom-configured broker
var _ = Describe("MSSQL Server and DB", Label("mssql-db"), func() {
	It("can be accessed by an app", func() {
		serverConfig := newDatabaseServer()

		By("Create CSB with server details")
		serverTag := random.Name(random.WithMaxLength(10))
		creds := serverConfig.getMASBServerDetails(serverTag)

		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-db"),
			brokers.WithLatestEnv(),
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: creds}),
		)
		defer serviceBroker.Delete()

		By("creating a server")
		dbs := mssqlserver.DatabaseServer{Name: serverConfig.Name, ResourceGroup: metadata.ResourceGroup}
		ctx := context.Background()
		Expect(mssqlserver.CreateResourceGroup(ctx, metadata.ResourceGroup, subscriptionID))
		Expect(mssqlserver.CreateServer(ctx, dbs, serverConfig.Username, serverConfig.Password, subscriptionID)).NotTo(HaveOccurred())
		defer func() {
			By("deleting the server")
			_ = mssqlserver.CleanupServer(ctx, dbs, subscriptionID)
		}()

		Expect(mssqlserver.CreateFirewallRule(ctx, metadata, dbs, subscriptionID)).NotTo(HaveOccurred())
		defer func() {
			By("deleting the firewall rule")
			_ = mssqlserver.CleanFirewallRule(ctx, dbs, subscriptionID)
		}()

		By("creating a database in the server")
		const serviceOffering = "csb-azure-mssql-db"
		const servicePlan = "small"
		serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
		// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
		// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
		// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
		// (which can only have 5 instances at any time) to prevent subsequent test failures.
		defer services.Delete(serviceName)
		dbInstance := services.CreateInstance(
			serviceOffering,
			servicePlan,
			services.WithBroker(serviceBroker),
			services.WithParameters(map[string]string{"server": serverTag}),
			services.WithName(serviceName),
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
		appOne.PUTf("", "%s?dbo=false", schema)

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appOne.PUTf(value, "%s/%s", schema, key)

		By("getting the value using the second app")
		got := appTwo.GETf("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("deleting binding one the binding two keeps reading the value - object reassignment works")
		binding.Unbind()
		got = appTwo.GETf("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the second app")
		appTwo.DELETE(schema)
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
