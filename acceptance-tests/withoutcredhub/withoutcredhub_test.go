package withoutcredhub_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", func() {
	It("can be accessed by an app", func() {
		env := helpers.EnvVar{Name: "CH_CRED_HUB_URL", Value: ""}
		broker := helpers.CreateBroker(helpers.BrokerWithPrefix("csb-storage"), helpers.BrokerWithEnv(env))
		defer broker.Delete()

		By("creating a service instance")
		serviceInstance := helpers.CreateServiceFromBroker("csb-azure-storage-account", "standard", broker.Name)
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		app := helpers.AppPushUnstarted(apps.Storage)
		defer helpers.AppDelete(app)

		By("binding the app to the storage service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		helpers.AppStart(app)

		By("checking that the app environment does not a credhub reference for credentials")
		Expect(binding.Credential()).NotTo(matchers.HaveCredHubRef)

		By("creating a collection")
		collectionName := random.Name(random.WithPrefix("collection"))
		app.PUT("", collectionName)

		By("uploading a blob")
		blobName := random.Hexadecimal()
		blobData := random.Hexadecimal()
		app.PUT(blobData, "%s/%s", collectionName, blobName)

		By("downloading the blob")
		got := app.GET("%s/%s", collectionName, blobName)
		Expect(got).To(Equal(blobData))
	})
})
