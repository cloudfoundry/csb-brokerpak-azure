package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	eventHubsServiceName             = "csb-azure-eventhubs"
	eventHubsServiceID               = "40b751ac-624d-11ea-8354-f38aff407636"
	eventHubsServiceDisplayName      = "Deprecated - Event Hubs"
	eventHubsServiceDescription      = "Deprecated - Simple, secure, and scalable real-time data ingestion"
	eventHubsServiceDocumentationURL = "https://azure.microsoft.com/en-us/services/event-hubs/"
	eventHubsServiceSupportURL       = "https://azure.microsoft.com/en-us/support/options/"
)

var _ = Describe("Event Hubs", Label("Event Hubs"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, eventHubsServiceName)
		Expect(service.ID).To(Equal(eventHubsServiceID))
		Expect(service.Description).To(Equal(eventHubsServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "eventhubs", "Event Hubs", "Azure", "preview", "deprecated"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(eventHubsServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(eventHubsServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(eventHubsServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("basic"),
					ID:   Equal("3ac4fede-62ed-11ea-af59-cb26248cfe7b"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("standard"),
					ID:   Equal("57e330ee-62ed-11ea-825c-23c5737ad688"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(eventHubsServiceName, "basic", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(eventHubsServiceName, "basic", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, eventHubsServiceName, "basic", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})
	})
})
