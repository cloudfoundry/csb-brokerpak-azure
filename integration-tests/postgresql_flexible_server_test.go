package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	postgreSQLFlexibleServerServiceName             = "csb-azure-postgresql-flexible-server"
	postgreSQLFlexibleServerServiceID               = "d69dd4aa-e27e-490c-bdbf-c887563da27f"
	postgreSQLFlexibleServerServiceDisplayName      = "Deprecated - Azure Database for PostgreSQL - flexible server"
	postgreSQLFlexibleServerServiceDescription      = "Deprecated - Azure Database for PostgreSQL - flexible server"
	postgreSQLFlexibleServerServiceDocumentationURL = "https://learn.microsoft.com/en-gb/azure/postgresql/"
	postgreSQLFlexibleServerServiceSupportURL       = "https://learn.microsoft.com/en-gb/azure/postgresql/"
	postgreSQLFlexibleServerCustomPlanName          = "custom-test"
	postgreSQLFlexibleServerCustomPlanID            = "43c6e11b-ffda-4787-856f-399a531cea34"
)

var customPostgresPlans = []map[string]any{
	customPostgresPlan,
}

var customPostgresPlan = map[string]any{
	"name":        postgreSQLFlexibleServerCustomPlanName,
	"id":          postgreSQLFlexibleServerCustomPlanID,
	"description": "Default Postgres plan",
	"metadata": map[string]any{
		"displayName": "custom-sample",
	},
}

var _ = Describe("PostgreSQL Flexible Server", Label("PostgreSQL-flexible-server"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, postgreSQLFlexibleServerServiceName)
		Expect(service.ID).To(Equal(postgreSQLFlexibleServerServiceID))
		Expect(service.Description).To(Equal(postgreSQLFlexibleServerServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "postgresql", "postgres", "preview", "flexible server", "deprecated"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(postgreSQLFlexibleServerServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(postgreSQLFlexibleServerServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(postgreSQLFlexibleServerServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal(postgreSQLFlexibleServerCustomPlanName),
					ID:   Equal(postgreSQLFlexibleServerCustomPlanID),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				_, err := broker.Provision(postgreSQLFlexibleServerServiceName, postgreSQLFlexibleServerCustomPlanName, params)

				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"invalid location",
				map[string]any{"location": "-Asia-northeast1"},
				"location: Does not match pattern '^[a-z][a-z0-9]+$'",
			),
			Entry(
				"instance name minimum length is 3 characters",
				map[string]any{"instance_name": stringOfLen(2)},
				"instance_name: String length must be greater than or equal to 3",
			),
			Entry(
				"instance name maximum length is 63 characters",
				map[string]any{"instance_name": stringOfLen(64)},
				"instance_name: String length must be less than or equal to 63",
			),
			Entry(
				"instance name invalid characters",
				map[string]any{"instance_name": ".aaaaa"},
				"instance_name: Does not match pattern '^[a-z][a-z0-9-]+$'",
			),
			Entry(
				"database name maximum length is 98 characters",
				map[string]any{"db_name": stringOfLen(99)},
				"db_name: String length must be less than or equal to 64",
			),
			Entry(
				"storage minimum is 32 GB",
				map[string]any{"storage_gb": 5},
				"storage_gb: Must be greater than or equal to 32",
			),
			Entry(
				"storage maximum is 32767 GB",
				map[string]any{"storage_gb": 33000},
				"storage_gb: Must be less than or equal to 32767",
			),
			Entry(
				"resource group invalid characters",
				map[string]any{"resource_group": ".rrrr"},
				"resource_group: Does not match pattern '^[a-z][a-z0-9-]+$|^$'",
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(postgreSQLFlexibleServerServiceName, postgreSQLFlexibleServerCustomPlanName, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("labels", MatchKeys(IgnoreExtras, Keys{
						"pcf-instance-id":       Equal(instanceID),
						"pcf-organization-guid": Equal(""),
						"pcf-space-guid":        Equal(""),
					})),
					HaveKeyWithValue("azure_client_id", armClientID),
					HaveKeyWithValue("azure_tenant_id", armTenantID),
					HaveKeyWithValue("azure_subscription_id", armSubscriptionID),
					HaveKeyWithValue("azure_client_secret", armClientSecret),
					HaveKeyWithValue("postgres_version", BeNil()),
					HaveKeyWithValue("storage_gb", float64(32)),
					HaveKeyWithValue("instance_name", fmt.Sprintf("csb-postgresql-%s", instanceID)),
					HaveKeyWithValue("db_name", "vsbdb"),
					HaveKeyWithValue("location", "westus"),
					HaveKeyWithValue("resource_group", ""),
					HaveKeyWithValue("sku_name", ""),
					HaveKeyWithValue("allow_access_from_azure_services", true),
					HaveKeyWithValue("delegated_subnet_id", BeNil()),
					HaveKeyWithValue("private_dns_zone_id", BeNil()),
					HaveKeyWithValue("private_endpoint_subnet_id", ""),
				),
			)
		})

		It("should allow properties to be set", func() {
			_, err := broker.Provision(postgreSQLFlexibleServerServiceName, postgreSQLFlexibleServerCustomPlanName, map[string]any{
				"instance_name":                    "test-instance",
				"db_name":                          "test-db",
				"storage_gb":                       64,
				"postgres_version":                 "14",
				"location":                         "eastus",
				"resource_group":                   "test-group",
				"sku_name":                         "GP_Standard_D2ads_v5",
				"allow_access_from_azure_services": false,
				"delegated_subnet_id":              "test-delegated-subnet",
				"private_dns_zone_id":              "test-private-dns-zone",
				"private_endpoint_subnet_id":       "test-private-endpoint-subnet-id",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("instance_name", "test-instance"),
					HaveKeyWithValue("postgres_version", "14"),
					HaveKeyWithValue("storage_gb", float64(64)),
					HaveKeyWithValue("db_name", "test-db"),
					HaveKeyWithValue("location", "eastus"),
					HaveKeyWithValue("resource_group", "test-group"),
					HaveKeyWithValue("sku_name", "GP_Standard_D2ads_v5"),
					HaveKeyWithValue("allow_access_from_azure_services", false),
					HaveKeyWithValue("delegated_subnet_id", "test-delegated-subnet"),
					HaveKeyWithValue("private_dns_zone_id", "test-private-dns-zone"),
					HaveKeyWithValue("private_endpoint_subnet_id", "test-private-endpoint-subnet-id"),
				),
			)
		})
	})

	Describe("updating", func() {
		var instanceID string

		BeforeEach(func() {
			var err error
			instanceID, err = broker.Provision(postgreSQLFlexibleServerServiceName, postgreSQLFlexibleServerCustomPlanName, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("should prevent updating properties flagged as `prohibit_update` because it can result in the recreation of the service instance",
			func(prop string, value any) {
				err := broker.Update(instanceID, postgreSQLFlexibleServerServiceName, postgreSQLFlexibleServerCustomPlanName, map[string]any{prop: value})

				Expect(err).To(MatchError(
					ContainSubstring(
						"attempt to update parameter that may result in service instance re-creation and data loss",
					),
				))

				const initialProvisionInvocation = 1
				Expect(mockTerraform.ApplyInvocations()).To(HaveLen(initialProvisionInvocation))
			},
			Entry("update location", "location", "no-matter-what-region"),
			Entry("update instance_name", "instance_name", "no-matter-what-name"),
			Entry("update resource_group", "resource_group", "no-matter-what-resource"),
			Entry("delegated_subnet_id", "delegated_subnet_id", "new-subnet"),
		)
	})

	Describe("binding", func() {
		var instanceID string
		BeforeEach(func() {
			err := mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "hostname", Type: "string", Value: "create.hostname.azure.test"},
				{Name: "username", Type: "string", Value: "create.test.username"},
				{Name: "password", Type: "string", Value: "create.test.password"},
				{Name: "name", Type: "string", Value: "create.test.instancename"},
				{Name: "port", Type: "number", Value: 5432},
			})
			Expect(err).NotTo(HaveOccurred())

			instanceID, err = broker.Provision(postgreSQLFlexibleServerServiceName, postgreSQLFlexibleServerCustomPlanName, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the bind values from terraform output", func() {
			err := mockTerraform.SetTFState([]testframework.TFStateValue{
				{Name: "username", Type: "string", Value: "bind.test.username"},
				{Name: "password", Type: "string", Value: "bind.test.password"},
				{Name: "uri", Type: "string", Value: "bind.test.uri"},
				{Name: "jdbcUrl", Type: "string", Value: "bind.test.jdbcUrl"},
			})
			Expect(err).NotTo(HaveOccurred())

			bindResult, err := broker.Bind(postgreSQLFlexibleServerServiceName, postgreSQLFlexibleServerCustomPlanName, instanceID, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(bindResult).To(Equal(map[string]any{
				"name":     "create.test.instancename",
				"hostname": "create.hostname.azure.test",
				"username": "bind.test.username",
				"password": "bind.test.password",
				"uri":      "bind.test.uri",
				"jdbcUrl":  "bind.test.jdbcUrl",
				"port":     float64(5432),
			}))
		})
	})
})
