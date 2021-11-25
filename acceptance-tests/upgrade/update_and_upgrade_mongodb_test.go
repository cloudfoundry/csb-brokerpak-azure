package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMongoTest", func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := helpers.CreateBroker(
				helpers.BrokerWithPrefix("csb-mssql-db"),
				helpers.BrokerFromDir(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service instance")
			databaseName := helpers.RandomName("database")
			collectionName := helpers.RandomName("collection")
			serviceInstance := helpers.CreateServiceFromBroker("csb-azure-mongodb", "small", serviceBroker.Name, map[string]interface{}{
				"db_name":         databaseName,
				"collection_name": collectionName,
				"shard_key":       "_id",
			})
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.MongoDB)
			appTwo := helpers.AppPushUnstarted(apps.MongoDB)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding the apps to the MongoDB service instance")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)

			By("starting the apps")
			helpers.AppStart(appOne, appTwo)

			By("creating a document using the first app")
			documentNameOne := helpers.RandomHex()
			documentDataOne := helpers.RandomHex()
			appOne.PUT(documentDataOne, "%s/%s/%s", databaseName, collectionName, documentNameOne)

			By("getting the document using the second app")
			got := appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("creating a document using the first app - post upgrade")
			documentNameTwo := helpers.RandomHex()
			documentDataTwo := helpers.RandomHex()
			appOne.PUT(documentDataTwo, "%s/%s/%s", databaseName, collectionName, documentNameTwo)

			By("getting the document using the second app")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameTwo)
			Expect(got).To(Equal(documentDataTwo))

		})
	})
})
