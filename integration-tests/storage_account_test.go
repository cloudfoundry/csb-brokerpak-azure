package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	storageAccountServiceName             = "csb-azure-storage-account"
	storageAccountServiceID               = "eb263d40-3a2e-4af1-9333-752acb1e6ea3"
	storageAccountServiceDisplayName      = "Deprecated - Azure Storage Account"
	storageAccountServiceDescription      = "Deprecated - Azure Storage Account"
	storageAccountServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/storage/common/storage-account-overview"
	storageAccountServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/storage/common/storage-account-overview"
)

var _ = Describe("Storage Account", Label("Storage Account"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, storageAccountServiceName)
		Expect(service.ID).To(Equal(storageAccountServiceID))
		Expect(service.Description).To(Equal(storageAccountServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "storage", "Azure", "preview", "Storage", "deprecated"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(storageAccountServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(storageAccountServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(storageAccountServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("standard"),
					ID:   Equal("b9fe2b0c-1a95-4a1b-a576-60e7f9e42aad"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(storageAccountServiceName, "standard", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(storageAccountServiceName, "standard", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, storageAccountServiceName, "standard", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})
	})
})
