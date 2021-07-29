package mongodb_test

import (
	"acceptancetests/helpers"
	"encoding/json"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MongoDB", func() {
	var (
		serviceInstanceName string
		databaseName        string
		collectionName      string
	)

	BeforeEach(func() {
		serviceInstanceName = helpers.RandomName("mongodb")
		databaseName = helpers.RandomName("database")
		collectionName = helpers.RandomName("collection")
		params, err := json.Marshal(map[string]interface{}{
			"db_name":         databaseName,
			"collection_name": collectionName,
			"shard_key":       "_id",
		})
		Expect(err).NotTo(HaveOccurred())
		helpers.CreateService("csb-azure-mongodb", "small", serviceInstanceName, string(params))
	})

	AfterEach(func() {
		helpers.DeleteService(serviceInstanceName)
	})

	It("can be accessed by an app", func() {
		By("building the app")
		appDir := helpers.AppBuild("./mongodbapp")
		defer os.RemoveAll(appDir)

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted("mongodb", appDir)
		appTwo := helpers.AppPushUnstarted("mongodb", appDir)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the MongoDB service instance")
		bindingName := helpers.Bind(appOne, serviceInstanceName)
		helpers.Bind(appTwo, serviceInstanceName)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		creds := helpers.GetBindingCredential(appOne, "csb-azure-mongodb", bindingName)
		Expect(creds).To(HaveKey("credhub-ref"))

		appOneURL := fmt.Sprintf("http://%s.%s", appOne, helpers.DefaultSharedDomain())
		By("checking that the specified database has been created")
		databases := helpers.HTTPGet(appOneURL)
		Expect(databases).To(MatchJSON(fmt.Sprintf(`["%s"]`, databaseName)))

		By("checking that the specified collection has been created")
		collections := helpers.HTTPGet(fmt.Sprintf("%s/%s", appOneURL, databaseName))
		Expect(collections).To(MatchJSON(fmt.Sprintf(`["%s"]`, collectionName)))

		By("creating a document using the first app")
		documentName := helpers.RandomString()
		documentData := helpers.RandomString()
		helpers.HTTPPost(fmt.Sprintf("%s/%s/%s/%s", appOneURL, databaseName, collectionName, documentName), documentData)

		By("getting the document using the second app")
		got := helpers.HTTPGet(fmt.Sprintf("http://%s.%s/%s/%s/%s", appTwo, helpers.DefaultSharedDomain(), databaseName, collectionName, documentName))
		Expect(got).To(Equal(documentData))
	})
})
