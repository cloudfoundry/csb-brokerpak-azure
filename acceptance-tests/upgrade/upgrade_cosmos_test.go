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
			documentNameOne := helpers.RandomHex()
			documentDataOne := helpers.RandomHex()
			appOne.PUT(documentDataOne, "%s/%s/%s", databaseName, collectionName, documentNameOne)

			By("getting the value using the second app")
			got := appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("creating a document using the first app - after upgrade")
			documentNameTwo := helpers.RandomHex()
			documentDataTwo := helpers.RandomHex()
			appOne.PUT(documentDataTwo, "%s/%s/%s", databaseName, collectionName, documentNameTwo)

			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameTwo)
			Expect(got).To(Equal(documentDataTwo))

			By("getting the value before broker upgrade")
			Expect(appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)).To(Equal(documentDataOne))

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("checking previous data still accessible")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameTwo)
			Expect(got).To(Equal(documentDataTwo))

			By("checking new data can be written and read")
			documentNameThree := helpers.RandomHex()
			documentDataThree := helpers.RandomHex()
			appOne.PUT(documentDataThree, "%s/%s/%s", databaseName, collectionName, documentNameThree)

			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameThree)
			Expect(got).To(Equal(documentDataThree))
		})
	})
})