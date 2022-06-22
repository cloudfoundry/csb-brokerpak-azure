package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MongoDB", Label("MongoDB"), func() {
	const serviceName = "csb-azure-mongodb"

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
	})
})
