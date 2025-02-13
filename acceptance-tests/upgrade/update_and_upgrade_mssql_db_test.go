package upgrade_test

import (
	"context"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/mssqlserver"
	"csbbrokerpakazure/acceptance-tests/helpers/plans"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

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
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serverConfig := newDatabaseServer()
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

			By("reconfiguring the CSB with DB server details")
			serverTag := random.Name(random.WithMaxLength(10))
			serviceBroker.UpdateEnv(apps.EnvVar{Name: "MSSQL_DB_SERVER_CREDS", Value: serverConfig.serverDetails(serverTag)})

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
			appOne.PUTf(valueOne, "%s/%s", schema, keyOne)

			By("getting the value using the second app")
			got := appTwo.GETf("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable(servicePlan, serviceOffering, serviceBroker.Name))

			By("upgrading service instance")
			dbInstance.Upgrade()

			By("checking previously created data still accessible")
			got = appTwo.GETf("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()

			By("creating new bindings")
			dbInstance.Bind(appOne)
			apps.Restage(appOne)

			By("updating service instance")
			dbInstance.Update("-c", `{}`)

			By("checking previously created data still accessible")
			got = appTwo.GETf("%s/%s", schema, keyOne)
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
			appOne.PUTf(valueTwo, "%s/%s", schema, keyTwo)

			got = appTwo.GETf("%s/%s", schema, keyTwo)
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
