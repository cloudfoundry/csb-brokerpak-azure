package storage_test

import (
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", func() {
	var (
		serviceInstance helpers.ServiceInstance
		collectionName  string
	)

	BeforeEach(func() {
		collectionName = helpers.RandomName("collection")
		serviceInstance = helpers.CreateService("csb-azure-storage-account", "standard")
	})

	AfterEach(func() {
		serviceInstance.Delete()
	})

	It("can be accessed by an app", func() {
		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted("storage", "./storageapp")
		appTwo := helpers.AppPushUnstarted("storage", "./storageapp")
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
		blobName := helpers.RandomString()
		blobData := helpers.RandomString()
		appOne.PUT(blobData, "%s/%s", collectionName, blobName)

		By("downloading the blob using the second app")
		got := appTwo.GET("%s/%s", collectionName, blobName)
		Expect(got).To(Equal(blobData))
	})
})
