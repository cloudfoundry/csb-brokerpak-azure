package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeStorageTest", func() {
	Context("When upgrading broker version", func(){
		It("should continue to work", func() {
			By("pushing latest released broker version")
			brokerName := helpers.RandomName("csb-storage")
			serviceBroker := helpers.PushAndStartBroker(brokerName, releasedBuildDir)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := helpers.CreateServiceFromBroker("csb-azure-storage-account", "standard", brokerName)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.Storage)
			appTwo := helpers.AppPushUnstarted(apps.Storage)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding to the apps")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppStart(appOne, appTwo)

			By("creating a collection")
			collectionName := helpers.RandomName("collection")
			appOne.PUT("", collectionName)

			By("uploading a blob using the first app")
			blobNameOne := helpers.RandomHex()
			blobDataOne := helpers.RandomHex()
			appOne.PUT(blobDataOne, "%s/%s", collectionName, blobNameOne)

			By("downloading the blob using the second app")
			got := appTwo.GET("%s/%s", collectionName, blobNameOne)
			Expect(got).To(Equal(blobDataOne))

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("checking that previously written data is accessible")
			got = appTwo.GET("%s/%s", collectionName, blobNameOne)
			Expect(got).To(Equal(blobDataOne))

			By("checking that data can still be written and read")
			blobNameTwo := helpers.RandomHex()
			blobDataTwo := helpers.RandomHex()
			appOne.PUT(blobDataTwo, "%s/%s", collectionName, blobNameTwo)
			got = appTwo.GET("%s/%s", collectionName, blobNameTwo)
			Expect(got).To(Equal(blobDataTwo))
		})
	})
})