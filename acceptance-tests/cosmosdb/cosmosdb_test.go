package cosmosdb_test

import (
	"acceptancetests/helpers"
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CosmosDB", func() {
	var (
		serviceInstanceName string
		databaseName        string
		collectionName      string
	)

	BeforeEach(func() {
		serviceInstanceName = helpers.RandomName("cosmosdb")
		databaseName = helpers.RandomName("database")
		collectionName = helpers.RandomName("collection")
		params, err := json.Marshal(map[string]interface{}{
			"db_name": databaseName,
		})
		Expect(err).NotTo(HaveOccurred())
		helpers.CreateService("csb-azure-cosmosdb-sql", "small", serviceInstanceName, string(params))
	})

	AfterEach(func() {
		helpers.DeleteService(serviceInstanceName)
	})

	It("can be accessed by an app", func() {
		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted("cosmosdb", "./cosmosdbapp")
		appTwo := helpers.AppPushUnstarted("cosmosdb", "./cosmosdbapp")
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the CosmosDB service instance")
		bindingName := helpers.Bind(appOne, serviceInstanceName)
		helpers.Bind(appTwo, serviceInstanceName)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		creds := helpers.GetBindingCredential(appOne, "csb-azure-cosmosdb-sql", bindingName)
		Expect(creds).To(HaveKey("credhub-ref"))

		appOneURL := fmt.Sprintf("http://%s.%s", appOne, helpers.DefaultSharedDomain())
		By("checking that the specified database has been created")
		databases := helpers.HTTPGet(appOneURL)
		Expect(databases).To(MatchJSON(fmt.Sprintf(`["%s"]`, databaseName)))

		By("creating a collection")
		helpers.HTTPPostJSON(
			fmt.Sprintf("%s/%s", appOneURL, databaseName),
			map[string]interface{}{"id": collectionName},
		)

		By("creating a document using the first app")
		documentName := helpers.RandomString()
		documentData := helpers.RandomString()
		helpers.HTTPPostJSON(
			fmt.Sprintf("%s/%s/%s", appOneURL, databaseName, collectionName),
			map[string]interface{}{"name": documentName, "data": documentData},
		)

		By("getting the document using the second app")
		got := helpers.HTTPGet(fmt.Sprintf("http://%s.%s/%s/%s/%s", appTwo, helpers.DefaultSharedDomain(), databaseName, collectionName, documentName))
		Expect(got).To(Equal(documentData))
	})
})
