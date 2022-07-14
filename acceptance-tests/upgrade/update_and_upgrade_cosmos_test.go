package upgrade_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeCosmosTest", Label("cosmosdb"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-cosmos"),
				brokers.WithSourceDir(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			databaseName := random.Name(random.WithPrefix("database"))
			serviceInstance := services.CreateInstance(
				"csb-azure-cosmosdb-sql",
				"small",
				services.WithBroker(serviceBroker),
				services.WithParameters(map[string]any{"db_name": databaseName}),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.Cosmos))
			appTwo := apps.Push(apps.WithApp(apps.Cosmos))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

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
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("checking previous data still accessible")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("updating the instance plan")
			serviceInstance.Update("-p", "medium")

			By("checking previous data still accessible")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameOne)
			Expect(got).To(Equal(documentDataOne))

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("creating a document using the first app - after upgrade")
			documentNameTwo := random.Hexadecimal()
			documentDataTwo := random.Hexadecimal()
			appOne.PUT(documentDataTwo, "%s/%s/%s", databaseName, collectionName, documentNameTwo)

			By("getting the document using the second app")
			got = appTwo.GET("%s/%s/%s", databaseName, collectionName, documentNameTwo)
			Expect(got).To(Equal(documentDataTwo))
		})
	})
})
