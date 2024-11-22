package upgrade_test

import (
	"fmt"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/az"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func updateMongoFirewall(serviceName, resourceGroup, publicIP string) {
	az.Start("cosmosdb", "update", "--ip-range-filter", publicIP, "--name", serviceName, "--resource-group", resourceGroup)
}

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
			databaseName := random.Name(random.WithPrefix("database"))
			collectionName := random.Name(random.WithPrefix("collection"))
			serviceInstance := services.CreateInstance(
				"csb-azure-mongodb",
				"small",
				services.WithBroker(serviceBroker),
				services.WithParameters(map[string]any{
					"db_name":         databaseName,
					"collection_name": collectionName,
					"shard_key":       "_id",
					"indexes":         "_id",
					"unique_indexes":  "",
				}),
			)
			defer serviceInstance.Delete()

			By("changing the firewall to allow comms")
			serviceName := fmt.Sprintf("csb%s", serviceInstance.GUID())
			resourceGroupName := fmt.Sprintf("rg-csb-mongo-%s", serviceInstance.GUID())
			updateMongoFirewall(serviceName, resourceGroupName, metadata.PublicIP)

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
			appOne.PUT(documentDataOne, "%s/%s/%s", databaseName, collectionName, documentNameOne)

			By("getting the document using the second app")
			got := appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("changing the firewall to allow comms")
			updateMongoFirewall(serviceName, resourceGroupName, metadata.PublicIP)

			By("checking previous data still accessible")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()

			By("creating new bindings")
			serviceInstance.Bind(appOne)
			apps.Restage(appOne)

			By("updating service instance")
			serviceInstance.Update("-c", `{}`)

			By("changing the firewall to allow comms")
			updateMongoFirewall(serviceName, resourceGroupName, metadata.PublicIP)

			By("checking previous data still accessible")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking previous data still accessible")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("creating a document using the first app - post upgrade")
			documentNameTwo := random.Hexadecimal()
			documentDataTwo := random.Hexadecimal()
			appOne.PUT(documentDataTwo, "%s/%s/%s", databaseName, collectionName, documentNameTwo)

			By("getting the document using the second app")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameTwo)
			Expect(got).To(Equal(documentDataTwo))
		})
	})
})
