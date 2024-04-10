package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	redisServiceName             = "csb-azure-redis"
	redisServiceID               = "349d89ac-2051-468b-b10f-9f537cc580c0"
	redisServiceDisplayName      = "Azure Cache for Redis"
	redisServiceDescription      = "Redis is a fully managed service for the Azure Platform"
	redisServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/azure-cache-for-redis/"
	redisServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/azure-cache-for-redis/"
)

var _ = Describe("Redis", Label("Redis"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, redisServiceName)
		Expect(service.ID).To(Equal(redisServiceID))
		Expect(service.Description).To(Equal(redisServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "redis", "preview"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(redisServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(redisServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(redisServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("deprecated-small"),
					ID:   Equal("6b9ca24e-1dec-4e6f-8c8a-dc6e11ab5bef"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("deprecated-medium"),
					ID:   Equal("6b272c43-2116-4483-9a99-de9262c0a7d6"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("deprecated-large"),
					ID:   Equal("c3e34abc-a820-457c-b723-1c342ef42c50"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("deprecated-ha-small"),
					ID:   Equal("d27a8e60-3724-49d1-b668-44b03d99b3b3"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("deprecated-ha-medium"),
					ID:   Equal("421b932a-b86f-48a3-97e4-64bb13d3ec13"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("deprecated-ha-large"),
					ID:   Equal("e919b281-9661-465d-82cf-0a0a6e1f195a"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("deprecated-ha-P1"),
					ID:   Equal("2a63e092-ab5c-4804-abd6-2d951240f0f6"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(redisServiceName, "deprecated-small", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(redisServiceName, "deprecated-small", map[string]any{})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-redis-%s", instanceID)),
					HaveKeyWithValue("location", "westus"),
					HaveKeyWithValue("sku_name", "Basic"),
					HaveKeyWithValue("family", "C"),
					HaveKeyWithValue("capacity", BeNumerically("==", 1)),
					HaveKeyWithValue("tls_min_version", "1.2"),
					HaveKeyWithValue("firewall_rules", []any{}),
					HaveKeyWithValue("private_endpoint_subnet_id", BeNil()),
					HaveKeyWithValue("private_dns_zone_ids", BeNil()),
					HaveKeyWithValue("resource_group", ""),
					HaveKeyWithValue("subnet_id", ""),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
				))
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(redisServiceName, "deprecated-small", map[string]any{
				"instance_name":              "test-instance",
				"resource_group":             "test-resource-group",
				"subnet_id":                  "test-subnet-id",
				"location":                   "centralus",
				"skip_provider_registration": true,
				"maxmemory_policy":           "some_other_policy",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("instance_name", "test-instance"),
					HaveKeyWithValue("resource_group", "test-resource-group"),
					HaveKeyWithValue("subnet_id", "test-subnet-id"),
					HaveKeyWithValue("location", "centralus"),
					HaveKeyWithValue("skip_provider_registration", true),
					HaveKeyWithValue("maxmemory_policy", "some_other_policy"),
				))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(redisServiceName, "deprecated-small", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, redisServiceName, "deprecated-small", map[string]any{"location": "asia-southeast1"})

			Expect(err).To(MatchError(
				ContainSubstring(
					"attempt to update parameter that may result in service instance re-creation and data loss",
				),
			))
			Expect(mockTerraform.ApplyInvocations()).To(HaveLen(0))
		})

		DescribeTable(
			"allowed updates",
			func(prop string, value any) {
				Expect(broker.Update(instanceID, redisServiceName, "deprecated-small", map[string]any{prop: value})).To(Succeed())
			},
			Entry("maxmemory_policy", "maxmemory_policy", "some_other_policy"),
		)
	})
})
