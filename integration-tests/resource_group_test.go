package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	resourceGroupServiceName             = "csb-azure-resource-group"
	resourceGroupServiceID               = "57af72ea-b951-44cb-b814-1da900554ce8"
	resourceGroupServiceDisplayName      = "Deprecated - Azure Resource Group"
	resourceGroupServiceDescription      = "Deprecated - Azure Resource Group"
	resourceGroupServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/overview#resource-groups"
	resourceGroupServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/overview#resource-groups"
)

var _ = Describe("Resource Group", Label("Resource Group"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, resourceGroupServiceName)
		Expect(service.ID).To(Equal(resourceGroupServiceID))
		Expect(service.Description).To(Equal(resourceGroupServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "preview", "deprecated"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(resourceGroupServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(resourceGroupServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(resourceGroupServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("standard"),
					ID:   Equal("c995c72a-a5f4-48d8-9179-9e295cc535b7"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(resourceGroupServiceName, "standard", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(resourceGroupServiceName, "standard", map[string]any{"instance_name": "resource-group-name"})

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, resourceGroupServiceName, "standard", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})
	})
})
