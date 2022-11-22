package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Redis", Label("Redis"), func() {
	const serviceName = "csb-azure-redis"

	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(serviceName, "small", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(serviceName, "small", map[string]any{})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-redis-%s", instanceID)),
					HaveKeyWithValue("location", "westus"),
					HaveKeyWithValue("sku_name", "Basic"),
					HaveKeyWithValue("family", "C"),
					HaveKeyWithValue("redis_version", "4"),
					HaveKeyWithValue("capacity", 1),
					HaveKeyWithValue("tls_min_version", "1.2"),
					HaveKeyWithValue("firewall_rules", ""),
					HaveKeyWithValue("private_endpoint_subnet_id", ""),
					HaveKeyWithValue("private_dns_zone_ids", ""),
					HaveKeyWithValue("resource_group", ""),
					HaveKeyWithValue("subnet_id", ""),
					HaveKeyWithValue("labels", HaveKeyWithValue("pcf-instance-id", instanceID)),
				))
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(serviceName, "small", map[string]any{
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
			instanceID, err = broker.Provision(serviceName, "small", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, serviceName, "small", map[string]any{"location": "asia-southeast1"})

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
				Expect(broker.Update(instanceID, serviceName, "small", map[string]any{prop: value})).To(Succeed())
			},
			Entry("maxmemory_policy", "maxmemory_policy", "some_other_policy"),
		)
	})
})
