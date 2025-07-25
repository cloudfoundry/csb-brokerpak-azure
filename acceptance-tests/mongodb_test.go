package acceptance_test

import (
	"fmt"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/az"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Tests the *csb-azure-mongodb* service offering
// Uses the *default broker*
var _ = Describe("MongoDB", Label("mongodb"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		databaseName := random.Name(random.WithPrefix("database"))
		collectionName := random.Name(random.WithPrefix("collection"))
		serviceInstance := services.CreateInstance(
			"csb-azure-mongodb",
			"small", services.WithParameters(map[string]any{
				"db_name":         databaseName,
				"collection_name": collectionName,
				"shard_key":       "_id",
				"indexes":         "_id",
				"unique_indexes":  "",
				"server_version":  "4.0",
			}),
		)
		defer serviceInstance.Delete()

		By("changing the firewall to allow comms")
		serviceName := fmt.Sprintf("csb%s", serviceInstance.GUID())
		updateMongoDBRangeFilter(serviceName)

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.MongoDB))
		appTwo := apps.Push(apps.WithApp(apps.MongoDB))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the MongoDB service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("checking that the specified database has been created")
		databases := appOne.GET("")
		Expect(databases).To(MatchJSON(fmt.Sprintf(`["%s"]`, databaseName)))

		By("checking that the specified collection has been created")
		collections := appOne.GET(databaseName)
		Expect(collections).To(MatchJSON(fmt.Sprintf(`["%s"]`, collectionName)))

		By("creating a document using the first app")
		documentName := random.Hexadecimal()
		documentData := random.Hexadecimal()
		appOne.PUTf(documentData, "%s/%s/%s", databaseName, collectionName, documentName)

		By("getting the document using the second app")
		got := appTwo.GETf("%s/%s/%s", databaseName, collectionName, documentName)
		Expect(got).To(Equal(documentData))
	})
})

func updateMongoDBRangeFilter(serviceName string) {
	var filter string
	switch {
	case firewallCIDR != "":
		GinkgoWriter.Println("Using specified firewall CIDR")
		filter = firewallCIDR
	case metadata.PublicIP != "":
		GinkgoWriter.Println("Using public IP from metadata")
		filter = metadata.PublicIP
	default:
		GinkgoWriter.Println("Not updating firewall")
		return
	}

	az.Run("cosmosdb", "update", "--ip-range-filter", filter, "--name", serviceName, "--resource-group", metadata.ResourceGroup)
}
