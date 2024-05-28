package integration_test

import (
	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	postgreSQLServiceName             = "csb-azure-postgresql"
	postgreSQLServiceID               = "ef89bc54-299a-4384-9dd6-4ea0cca11700"
	postgreSQLServiceDisplayName      = "Azure Database for PostgreSQL"
	postgreSQLServiceDescription      = "Azure Database for PostgreSQL"
	postgreSQLServiceDocumentationURL = "https://azure.microsoft.com/en-us/services/postgresql/"
	postgreSQLServiceSupportURL       = "https://azure.microsoft.com/en-us/services/postgresql/"
)

var _ = Describe("PostgreSQL Single Server", Label("PostgreSQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, postgreSQLServiceName)
		Expect(service.ID).To(Equal(postgreSQLServiceID))
		Expect(service.Description).To(Equal(postgreSQLServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "postgresql", "postgres", "preview"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(postgreSQLServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(postgreSQLServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(postgreSQLServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small"),
					ID:   Equal("aa1cf0f0-79fe-4132-a112-859fef9bf7cc"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("d83c74b9-0bb9-409e-bd26-8424ea908462"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("945b0ca1-5b5e-49b0-884a-20809103907e"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		It("should check location constraints", func() {
			_, err := broker.Provision(postgreSQLServiceName, "small", map[string]any{"location": "-Asia-northeast1"})

			Expect(err).To(MatchError(ContainSubstring("location: Does not match pattern '^[a-z][a-z0-9]+$'")))
		})
	})

	Describe("updating instance", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(postgreSQLServiceName, "small", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(mockTerraform.Reset()).To(Succeed())
		})

		It("should prevent updating location due to is flagged as `prohibit_update` and it can result in the recreation of the service instance and lost data", func() {
			err := broker.Update(instanceID, postgreSQLServiceName, "small", map[string]any{"location": "asia-southeast1"})

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

			instanceID, err := broker.Provision(postgreSQLServiceName, "small", nil)
			Expect(err).NotTo(HaveOccurred())

			err = mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "username", Type: "string", Value: "bind.test.username"},
				{Name: "password", Type: "string", Value: "bind.test.password"},
				{Name: "uri", Type: "string", Value: "bind.test.uri"},
				{Name: "jdbcUrl", Type: "string", Value: "bind.test.jdbcUrl"},
			})
			Expect(err).NotTo(HaveOccurred())
			bindResult, err := broker.Bind(postgreSQLServiceName, "small", instanceID, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(Equal(map[string]any{
				"username": "bind.test.username",
				"hostname": "create.hostname.azure.test",
				"jdbcUrl":  "bind.test.jdbcUrl",
				"name":     "create.test.instancename",
				"password": "bind.test.password",
				"uri":      "bind.test.uri",
				"use_tls":  true,
				"port":     float64(5443),
			}))
		})
	})
})
