package cosmosdb_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CosmosDB", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		databaseName := random.Name(random.WithPrefix("database"))
		serviceInstance := services.CreateInstance(
			"csb-azure-cosmosdb-sql",
			"small",
			services.WithParameters(map[string]interface{}{"db_name": databaseName}),
		)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.Cosmos))
		appTwo := apps.Push(apps.WithApp(apps.Cosmos))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the CosmosDB service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("checking that the specified database has been created")
		databases := appOne.GET("/")
		Expect(databases).To(MatchJSON(fmt.Sprintf(`["%s"]`, databaseName)))

		By("creating a collection")
		collectionName := random.Name(random.WithPrefix("collection"))
		appOne.PUT("", "%s/%s", databaseName, collectionName)

		By("creating a document using the first app")
		documentName := random.Hexadecimal()
		documentData := random.Hexadecimal()
		appOne.PUT(documentData, "%s/%s/%s", databaseName, collectionName, documentName)

		By("getting the document using the second app")
		got := appTwo.GET("%s/%s/%s", databaseName, collectionName, documentName)
		Expect(got).To(Equal(documentData))
	})
})
