package storage_test

import (
	"acceptancetests/helpers"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", func() {
	var (
		serviceInstanceName string
		collectionName      string
	)

	BeforeEach(func() {
		serviceInstanceName = helpers.RandomName("storage")
		collectionName = helpers.RandomName("collection")
		helpers.CreateService("csb-azure-storage-account", "standard", serviceInstanceName)
	})

	AfterEach(func() {
		helpers.DeleteService(serviceInstanceName)
	})

	It("can be accessed by an app", func() {
		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted("storage", "./storageapp")
		appTwo := helpers.AppPushUnstarted("storage", "./storageapp")
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the storage service instance")
		bindingName := helpers.Bind(appOne, serviceInstanceName)
		helpers.Bind(appTwo, serviceInstanceName)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		creds := helpers.GetBindingCredential(appOne, "csb-azure-storage-account", bindingName)
		Expect(creds).To(HaveKey("credhub-ref"))

		appOneURL := fmt.Sprintf("http://%s.%s", appOne, helpers.DefaultSharedDomain())
		By("creating a collection")
		helpers.HTTPPut(fmt.Sprintf("%s/%s", appOneURL, collectionName), "")

		By("uploading a blob using the first app")
		blobName := helpers.RandomString()
		blobData := helpers.RandomString()
		helpers.HTTPPut(fmt.Sprintf("%s/%s/%s", appOneURL, collectionName, blobName), blobData)

		By("downloading the blob using the second app")
		got := helpers.HTTPGet(fmt.Sprintf("http://%s.%s/%s/%s", appTwo, helpers.DefaultSharedDomain(), collectionName, blobName))
		Expect(got).To(Equal(blobData))
	})
})
