package cosmosdb_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CosmosDB", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		databaseName := helpers.RandomName("database")
		serviceInstance := helpers.CreateServiceInBroker(
			"csb-azure-cosmosdb-sql",
			"small",
			helpers.DefaultBroker().Name,
			map[string]interface{}{"db_name": databaseName })
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.Cosmos)
		appTwo := helpers.AppPushUnstarted(apps.Cosmos)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the CosmosDB service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

		By("checking that the specified database has been created")
		databases := appOne.GET("/")
		Expect(databases).To(MatchJSON(fmt.Sprintf(`["%s"]`, databaseName)))

		By("creating a collection")
		collectionName := helpers.RandomName("collection")
		appOne.PUT("", "%s/%s", databaseName, collectionName)

		By("creating a document using the first app")
		documentName := helpers.RandomHex()
		documentData := helpers.RandomHex()
		appOne.PUT(documentData, "%s/%s/%s", databaseName, collectionName, documentName)

		By("getting the document using the second app")
		got := appTwo.GET("%s/%s/%s", databaseName, collectionName, documentName)
		Expect(got).To(Equal(documentData))
	})
})
