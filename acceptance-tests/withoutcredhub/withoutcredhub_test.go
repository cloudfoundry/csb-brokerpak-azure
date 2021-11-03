package withoutcredhub_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Without CredHub", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := helpers.CreateServiceInBroker("csb-azure-storage-account", "standard", helpers.DefaultBroker().Name,)
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		app := helpers.AppPushUnstarted(apps.Storage)
		defer helpers.AppDelete(app)

		By("binding the app to the storage service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		helpers.AppStart(app)

		By("checking that the app environment does not a credhub reference for credentials")
		Expect(binding.Credential()).NotTo(helpers.HaveCredHubRef)

		By("creating a collection")
		collectionName := helpers.RandomName("collection")
		app.PUT("", collectionName)

		By("uploading a blob")
		blobName := helpers.RandomHex()
		blobData := helpers.RandomHex()
		app.PUT(blobData, "%s/%s", collectionName, blobName)

		By("downloading the blob")
		got := app.GET("%s/%s", collectionName, blobName)
		Expect(got).To(Equal(blobData))
	})
})
