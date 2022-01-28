package upgrade_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeStorageTest", Label("storage"), func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-storage"),
				brokers.WithSourceDir(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-azure-storage-account",
				"standard",
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.Storage))
			appTwo := apps.Push(apps.WithApp(apps.Storage))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)
			apps.Start(appOne, appTwo)

			By("creating a collection")
			collectionName := random.Name(random.WithPrefix("collection"))
			appOne.PUT("", collectionName)

			By("uploading a blob using the first app")
			blobNameOne := random.Hexadecimal()
			blobDataOne := random.Hexadecimal()
			appOne.PUT(blobDataOne, "%s/%s", collectionName, blobNameOne)

			By("downloading the blob using the second app")
			got := appTwo.GET("%s/%s", collectionName, blobNameOne)
			Expect(got).To(Equal(blobDataOne))

			By("pushing the development version of the broker")
			serviceBroker.UpdateSourceDir(developmentBuildDir)

			By("re-applying the terraform for service instance")
			serviceInstance.Update("-c", "{\"replication_type\":\"GRS\"}")

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("checking that previously written data is accessible")
			got = appTwo.GET("%s/%s", collectionName, blobNameOne)
			Expect(got).To(Equal(blobDataOne))

			By("checking that data can still be written and read")
			blobNameTwo := random.Hexadecimal()
			blobDataTwo := random.Hexadecimal()
			appOne.PUT(blobDataTwo, "%s/%s", collectionName, blobNameTwo)
			got = appTwo.GET("%s/%s", collectionName, blobNameTwo)
			Expect(got).To(Equal(blobDataTwo))
		})
	})
})
