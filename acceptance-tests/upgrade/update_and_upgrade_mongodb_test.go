package upgrade_test

import (
	"fmt"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/az"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/plans"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMongoTest", Label("mongodb"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-mongodb"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			const serviceOffering = "csb-azure-mongodb"
			const servicePlan = "small"
			serviceName := random.Name(random.WithPrefix(serviceOffering, servicePlan))
			// CreateInstance can fail and can leave a service record (albeit a failed one) lying around.
			// We can't delete service brokers that have serviceInstances, so we need to ensure the service instance
			// is cleaned up regardless as to whether it wa successful. This is important when we use our own service broker
			// (which can only have 5 instances at any time) to prevent subsequent test failures.
			defer services.Delete(serviceName)

			databaseName := random.Name(random.WithPrefix("database"))
			collectionName := random.Name(random.WithPrefix("collection"))
			serviceInstance := services.CreateInstance(
				serviceOffering,
				servicePlan,
				services.WithBroker(serviceBroker),
				services.WithParameters(map[string]any{
					"db_name":         databaseName,
					"collection_name": collectionName,
					"shard_key":       "_id",
					"indexes":         "_id",
					"unique_indexes":  "",
					"server_version":  "4.0",
				}),
				services.WithName(serviceName),
			)

			By("changing the firewall to allow comms")
			azureMongoResourceName := fmt.Sprintf("csb%s", serviceInstance.GUID())
			updateMongoDBRangeFilter(azureMongoResourceName)

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MongoDB))
			appTwo := apps.Push(apps.WithApp(apps.MongoDB))
			defer apps.Delete(appOne, appTwo)

			By("binding the apps to the MongoDB service instance")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

			By("creating a document using the first app")
			documentNameOne := random.Hexadecimal()
			documentDataOne := random.Hexadecimal()
			appOne.PUTf(documentDataOne, "%s/%s/%s", databaseName, collectionName, documentNameOne)

			By("getting the document using the second app")
			got := appTwo.GETf("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("validating that the instance plan is still active")
			Expect(plans.ExistsAndAvailable(servicePlan, serviceOffering, serviceBroker.Name))

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("changing the firewall to allow comms")
			updateMongoDBRangeFilter(azureMongoResourceName)

			By("checking previous data still accessible")
			got = appTwo.GETf("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()

			By("creating new bindings")
			serviceInstance.Bind(appOne)
			apps.Restage(appOne)

			By("updating service instance")
			serviceInstance.Update("-c", `{}`)

			By("changing the firewall to allow comms")
			updateMongoDBRangeFilter(azureMongoResourceName)

			By("checking previous data still accessible")
			got = appTwo.GETf("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking previous data still accessible")
			got = appTwo.GETf("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("creating a document using the first app - post upgrade")
			documentNameTwo := random.Hexadecimal()
			documentDataTwo := random.Hexadecimal()
			appOne.PUTf(documentDataTwo, "%s/%s/%s", databaseName, collectionName, documentNameTwo)

			By("getting the document using the second app")
			got = appTwo.GETf("%s/%s/%s", databaseName, collectionName, documentNameTwo)
			Expect(got).To(Equal(documentDataTwo))
		})
	})
})

func updateMongoDBRangeFilter(serviceName string) {
	var filter string
	switch {
	case firewallCIDR != "":
		GinkgoWriter.Println("Using specified firewall CIDR")
		filter = firewallCIDR
	case metadata.PublicIP != "":
		GinkgoWriter.Println("Using public IP from metadata")
		filter = metadata.PublicIP
	default:
		GinkgoWriter.Println("Not updating firewall")
		return
	}

	az.Run("cosmosdb", "update", "--ip-range-filter", filter, "--name", serviceName, "--resource-group", metadata.ResourceGroup)
}
