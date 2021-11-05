package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeCosmosTest", func() {
	Context("When upgrading broker version", func(){
		It("should continue to work", func() {
			By("pushing latest released broker version")
			brokerName := helpers.RandomName("csb-cosmos")
			serviceBroker := helpers.PushAndStartBroker(brokerName, releasedBuildDir)
			defer serviceBroker.Delete()

			By("creating a service")
			databaseName := helpers.RandomName("database")
			serviceInstance := helpers.CreateServiceFromBroker(
				"csb-azure-cosmosdb-sql",
				"small",
				brokerName,
				map[string]interface{}{"db_name": databaseName })
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
			collectionName := helpers.RandomName("collection")
			appOne.PUT("", "%s/%s", databaseName, collectionName)

			By("creating a document using the first app")
			documentName := helpers.RandomHex()
			documentData := helpers.RandomHex()
			appOne.PUT(documentData, "%s/%s/%s", databaseName, collectionName, documentName)

			By("getting the value using the second app")
			got := appTwo.GET("%s/%s/%s", databaseName, collectionName, documentName)
			Expect(got).To(Equal(documentData))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("creating a document using the first app")
			documentName = helpers.RandomHex()
			documentData = helpers.RandomHex()
			appOne.PUT(documentData, "%s/%s/%s", databaseName, collectionName, documentName)

			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentName)
			Expect(got).To(Equal(documentData))

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("checking it still works")
			documentName = helpers.RandomHex()
			documentData = helpers.RandomHex()
			appOne.PUT(documentData, "%s/%s/%s", databaseName, collectionName, documentName)

			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentName)
			Expect(got).To(Equal(documentData))
		})
	})
})