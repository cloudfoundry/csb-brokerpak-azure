package upgrade_test

import (
	"context"
	"net/http"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/mssqlserver"
	"csbbrokerpakazure/acceptance-tests/helpers/plans"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlDBFailoverGroupTest", Label("mssql-db-failover-group"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {

			ctx := context.Background()

			By("creating primary and secondary DB servers in their resource group")
			serversConfig, err := mssqlserver.CreateServerPair(ctx, metadata, subscriptionID)
			Expect(err).NotTo(HaveOccurred())

			DeferCleanup(func() {
				GinkgoHelper()

				By("deleting the created resource group and DB servers")
				err := mssqlserver.Cleanup(ctx, serversConfig, subscriptionID)
				Expect(err).NotTo(HaveOccurred())
			})

			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-db-fo"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
				brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
			)
			defer serviceBroker.Delete()

			By("creating a failover group service instance")
			const serviceOffering = "csb-azure-mssql-db-failover-group"
			const servicePlan = "small"
			serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
			// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
			// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
			// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
			// (which can only have 5 instances at any time) to prevent subsequent test failures.
			defer services.Delete(serviceName)
			fogConfig := newFailoverGroupConfig(serversConfig.ServerPairTag)
			initialFogInstance := services.CreateInstance(
				serviceOffering,
				servicePlan,
				services.WithBroker(serviceBroker),
				services.WithParameters(fogConfig),
				services.WithName(serviceName),
			)

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MSSQL))
			appTwo := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := initialFogInstance.Bind(appOne)
			bindingTwo := initialFogInstance.Bind(appTwo)

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

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable(servicePlan, serviceOffering, serviceBroker.Name))

			By("upgrading previous services")
			initialFogInstance.Upgrade()

			// When the "azurerm_sql_failover_group" resource is updated to the "azurerm_mssql_failover_group" resource,
			// the name of the FOG has to be changed (to avoid name clashes), resulting in the URL of the FOG
			// changing. This means that existing bindings do not work, and apps need to be re-bound.
			// At some stage this behavior will change, and this check and the subsequent re-bind can be removed.
			By("observing that existing bindings do not work")
			resp := appTwo.GETResponse("%s/%s", schema, keyOne)
			Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(resp).To(HaveHTTPBody(MatchRegexp(`error preparing statement:.*no such host`)))

			By("re-binding and re-staging the apps")
			bindingOne.Unbind()
			bindingTwo.Unbind()
			initialFogInstance.Bind(appOne)
			initialFogInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("updating the instance")
			initialFogInstance.Update("-c", "{}")

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			// We now append a "-g" to the instance name as we replace the "azurerm_sql_failover_group"
			// resource with the "azurerm_mssql_failover_group" to avoid name clashes.
			fogConfig.InstanceName += "-g"

			By("connecting to the existing failover group")
			const servicePlanExisting = "existing"
			serviceNameExisting := random.Name(random.WithPrefix(serviceOffering, servicePlanExisting))
			defer services.Delete(serviceNameExisting)
			dbFogInstance := services.CreateInstance(
				serviceOffering,
				servicePlanExisting,
				services.WithBroker(serviceBroker),
				services.WithParameters(fogConfig),
				services.WithName(serviceNameExisting),
			)

			By("purging the initial FOG instance")
			initialFogInstance.Purge()

			By("creating new bindings using the 'existing' plan")
			bindingOne = dbFogInstance.Bind(appOne)
			bindingTwo = dbFogInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)
			defer bindingOne.Unbind()
			defer bindingTwo.Unbind()

			By("getting the previously set values")
			Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

			By("checking data can be written and read")
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

type failoverGroupConfig struct {
	InstanceName  string `json:"instance_name"`
	DBName        string `json:"db_name"`
	ServerPairTag string `json:"server_pair"`
}

func newFailoverGroupConfig(serverPairTag string) failoverGroupConfig {
	return failoverGroupConfig{
		InstanceName:  random.Name(random.WithPrefix("fog")),
		DBName:        random.Name(random.WithPrefix("db")),
		ServerPairTag: serverPairTag,
	}
}
