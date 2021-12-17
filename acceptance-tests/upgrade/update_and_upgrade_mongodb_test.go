package upgrade_test

import (
	"acceptancetests/helpers"
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMongoTest", func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-mongodb"),
				brokers.WithSourceDir(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			databaseName := random.Name(random.WithPrefix("database"))
			collectionName := random.Name(random.WithPrefix("collection"))
			serviceInstance := helpers.CreateServiceFromBroker("csb-azure-mongodb", "small", serviceBroker.Name, map[string]interface{}{
				"db_name":         databaseName,
				"collection_name": collectionName,
				"shard_key":       "_id",
			})
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MongoDB))
			appTwo := apps.Push(apps.WithApp(apps.MongoDB))
			defer apps.Delete(appOne, appTwo)

			By("binding the apps to the MongoDB service instance")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)

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
			serviceBroker.UpdateSourceDir(developmentBuildDir)

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("checking previous data still accessible")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

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
