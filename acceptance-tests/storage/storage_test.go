package storage_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		collectionName := helpers.RandomName("collection")
		serviceInstance := helpers.CreateServiceInBroker("csb-azure-storage-account", "standard", helpers.DefaultBroker().Name)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.Storage)
		appTwo := helpers.AppPushUnstarted(apps.Storage)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the storage service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

		By("creating a collection")
		appOne.PUT("", collectionName)

		By("uploading a blob using the first app")
		blobName := helpers.RandomHex()
		blobData := helpers.RandomHex()
		appOne.PUT(blobData, "%s/%s", collectionName, blobName)

		By("downloading the blob using the second app")
		got := appTwo.GET("%s/%s", collectionName, blobName)
		Expect(got).To(Equal(blobData))
	})
})
