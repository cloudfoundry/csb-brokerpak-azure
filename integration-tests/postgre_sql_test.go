package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL", Label("PostgreSQL"), func() {
	const serviceName = "csb-azure-postgresql"

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

	Context("bind a service ", func() {
		It("return the bind values from terraform output", func() {
			err := mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "hostname", Type: "string", Value: "create.hostname.azure.test"},
				{Name: "username", Type: "string", Value: "create.test.username"},
				{Name: "password", Type: "string", Value: "create.test.password"},
				{Name: "name", Type: "string", Value: "create.test.instancename"},
				{Name: "use_tls", Type: "bool", Value: true},
				{Name: "port", Type: "number", Value: 5443},
			})
			Expect(err).NotTo(HaveOccurred())

			instanceID, err := broker.Provision(serviceName, "small", nil)
			Expect(err).NotTo(HaveOccurred())

			err = mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "username", Type: "string", Value: "bind.test.username"},
				{Name: "password", Type: "string", Value: "bind.test.password"},
				{Name: "uri", Type: "string", Value: "bind.test.uri"},
				{Name: "jdbcUrl", Type: "string", Value: "bind.test.jdbcUrl"},
			})
			Expect(err).NotTo(HaveOccurred())
			bindResult, err := broker.Bind(serviceName, "small", instanceID, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(Equal(map[string]any{
				"username":    "bind.test.username",
				"hostname":    "create.hostname.azure.test",
				"jdbcUrl":     "bind.test.jdbcUrl",
				"name":        "create.test.instancename",
				"password":    "bind.test.password",
				"uri":         "bind.test.uri",
				"require_ssl": true,
				"port": 5443,
			}))
		})
	})
})
