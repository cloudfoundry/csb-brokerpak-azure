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

// Tests the *csb-azure-mssql-db-failover-group* with the *existing* property that allows it to adopt and existing
// failover group rather than creating a new one.
// Does NOT use the default broker: deploys a custom-configured broker
var _ = Describe("MSSQL Failover Group Existing", Label("mssql-db-failover-group-existing"), func() {
	It("can be accessed by an app", func() {
		ctx := context.Background()

		By("creating primary and secondary DB servers in their resource group")
		serversConfig, err := mssqlserver.CreateServerPair(ctx, metadata, subscriptionID)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			By("deleting the created resource group and DB servers")
			_ = mssqlserver.Cleanup(ctx, serversConfig, subscriptionID)
		})

		By("deploying the CSB")

		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-db-fog"),
			brokers.WithLatestEnv(),
			brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
		)
		defer serviceBroker.Delete()

		By("creating a failover group service instance")
		fogConfig := map[string]string{
			"instance_name": random.Name(random.WithPrefix("fog")),
			"db_name":       random.Name(random.WithPrefix("db")),
			"server_pair":   serversConfig.ServerPairTag,
		}

		const serviceOffering = "csb-azure-mssql-db-failover-group"
		const servicePlan = "medium"
		serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
		// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
		// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
		// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
		// (which can only have 5 instances at any time) to prevent subsequent test failures.
		defer services.Delete(serviceName)
		initialFogInstance := services.CreateInstance(
			serviceOffering,
			servicePlan,
			services.WithBroker(serviceBroker),
			services.WithParameters(fogConfig),
			services.WithName(serviceName),
		)

		By("pushing an unstarted app")
		app := apps.Push(apps.WithApp(apps.MSSQL))

		By("binding the app to the initial failover group service instance")
		initialFogInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("creating a schema")
		schema := random.Name(random.WithMaxLength(10))
		app.PUTf("", "%s?dbo=false", schema)

		By("setting a key-value")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		app.PUTf(value, "%s/%s", schema, key)

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

		By("binding the app to the CSB service instance")
		bindingTwo := dbFogInstance.Bind(app)
		defer apps.Delete(app) // app needs to be deleted before service instance

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingTwo.Credential()).To(matchers.HaveCredHubRef)

		By("getting the value set with the initial binding")
		got := app.GETf("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the app")
		app.DELETE(schema)
	})
})
