package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/helpers/random"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeCosmosTest", func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := helpers.CreateBroker(
				helpers.BrokerWithPrefix("csb-cosmos"),
				helpers.BrokerFromDir(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			databaseName := random.Name(random.WithPrefix("database"))
			serviceInstance := helpers.CreateServiceFromBroker(
				"csb-azure-cosmosdb-sql",
				"small",
				serviceBroker.Name,
				map[string]interface{}{"db_name": databaseName})
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.Cosmos)
			appTwo := helpers.AppPushUnstarted(apps.Cosmos)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding to the apps")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)

			By("starting the apps")
			helpers.AppStart(appOne, appTwo)

			By("checking that the specified database has been created")
			databases := appOne.GET("/")
			Expect(databases).To(MatchJSON(fmt.Sprintf(`["%s"]`, databaseName)))
			databases = appTwo.GET("/")
			Expect(databases).To(MatchJSON(fmt.Sprintf(`["%s"]`, databaseName)))

			By("creating a collection")
			collectionName := random.Name(random.WithPrefix("collection"))
			appOne.PUT("", "%s/%s", databaseName, collectionName)

			By("creating a document using the first app")
			documentNameOne := random.Hexadecimal()
			documentDataOne := random.Hexadecimal()
			appOne.PUT(documentDataOne, "%s/%s/%s", databaseName, collectionName, documentNameOne)

			By("getting the value using the second app")
			got := appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("checking previous data still accessible")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("creating a document using the first app - after upgrade")
			documentNameTwo := random.Hexadecimal()
			documentDataTwo := random.Hexadecimal()
			appOne.PUT(documentDataTwo, "%s/%s/%s", databaseName, collectionName, documentNameTwo)

			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameTwo)
			Expect(got).To(Equal(documentDataTwo))
		})
	})
})
