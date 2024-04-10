package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mssqlDBServiceName             = "csb-azure-mssql-db"
	mssqlDBServiceID               = "6663f9f1-33c1-4f7d-839c-d4b7682d88cc"
	mssqlDBServiceDisplayName      = "Azure SQL Database"
	mssqlDBServiceDescription      = "Manage Azure SQL Databases on pre-provisioned database servers"
	mssqlDBServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/sql-database/"
	mssqlDBServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/sql-database/"
	msSQLDBCustomPlanName          = "custom-sample"
	msSQLDBCustomPlanID            = "d10e9572-0ea8-4bad-a4f3-a9a084dde067"
)

var customMSSQLDBPlans = []map[string]any{
	customMSSQLDBPlan,
}

var customMSSQLDBPlan = map[string]any{
	"name":        msSQLDBCustomPlanName,
	"id":          msSQLDBCustomPlanID,
	"description": "Default MSSQL DB plan",
	"metadata": map[string]any{
		"displayName": "custom-sample",
	},
}

var _ = Describe("MSSQL DB", Label("MSSQL"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mssqlDBServiceName)
		Expect(service.ID).To(Equal(mssqlDBServiceID))
		Expect(service.Description).To(Equal(mssqlDBServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "mssql", "sqlserver", "preview"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(mssqlDBServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(mssqlDBServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(mssqlDBServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("custom-sample"),
					ID:   Equal("d10e9572-0ea8-4bad-a4f3-a9a084dde067"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small"),
					ID:   Equal("fd07d12b-94f8-4f69-bd5b-e2c4e84fafc1"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("3ee14bce-33e8-4d02-9850-023a66bfe120"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("8f1c9c7b-80b2-49c3-9365-a3a059df9907"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("extra-large"),
					ID:   Equal("09096759-58a8-41d0-96bf-39b02a0e4104"),
				}),
			),
		)
	})

	Describe("provisioning", func() {

		DescribeTable("property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				params["server"] = "preprovisioned-server-name"
				_, err := broker.Provision(mssqlDBServiceName, msSQLDBCustomPlanName, params)
				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"db name maximum length is 64 characters",
				map[string]any{"db_name": stringOfLen(65)},
				"db_name: String length must be less than or equal to 64",
			),
			Entry(
				"cores maximum is 80",
				map[string]any{"cores": 82},
				"cores: Must be less than or equal to 80",
			),
			Entry(
				"cores must be even",
				map[string]any{"cores": 3},
				"cores: Must be a multiple of 2",
			),
			Entry(
				"cores must be greater than 0",
				map[string]any{"cores": 0},
				"cores: Must be greater than or equal to 1",
			),
			Entry(
				"max storage must be at least 1",
				map[string]any{"max_storage_gb": 0},
				"max_storage_gb: Must be greater than or equal to 1",
			),
			Entry(
				"short_term_retention_days must be less or equal 35",
				map[string]any{"short_term_retention_days": 36},
				"short_term_retention_days: Must be less than or equal to 35",
			),
			Entry(
				"ltr_week_of_year can't be 0",
				map[string]any{"ltr_week_of_year": 0},
				"ltr_week_of_year: Must be greater than or equal to 1",
			),
			Entry(
				"ltr_week_of_year can't be > 52",
				map[string]any{"ltr_week_of_year": 53},
				"ltr_week_of_year: Must be less than or equal to 52",
			),
			Entry(
				"ltr_weekly_retention invalid pattern",
				map[string]any{"ltr_weekly_retention": "PT010304"},
				"ltr_weekly_retention: Does not match pattern '^(P|PT)(?:[0-9]|[1-9][0-9]|[1-4][0-9]{2}|5[0-1][0-9]|520)(W|S)$'",
			),
			Entry(
				"ltr_monthly_retention invalid pattern",
				map[string]any{"ltr_monthly_retention": "PT0103N"},
				"ltr_monthly_retention: Does not match pattern '^(P|PT)([0-9]{1,2}|1[01][0-9]|12[0])(M|S)$'",
			),
			Entry(
				"ltr_yearly_retention invalid pattern",
				map[string]any{"ltr_yearly_retention": "PT0103N"},
				"ltr_yearly_retention: Does not match pattern '^(P|PT)10|[0-9](Y|S)$'",
			),
		)

		It("should provision a plan", func() {
			instanceID, err := broker.Provision(mssqlDBServiceName, customMSSQLDBPlan["name"].(string), map[string]any{"server": "preprovisioned-server-name"})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("cores", BeNumerically("==", 2)),
					HaveKeyWithValue("max_storage_gb", BeNumerically("==", 5)),
					HaveKeyWithValue("db_name", fmt.Sprintf("csb-db-%s", instanceID)),
					HaveKeyWithValue("server", "preprovisioned-server-name"),
					HaveKeyWithValue("sku_name", BeEmpty()),
					HaveKeyWithValue("short_term_retention_days", BeNumerically("==", 7)),
					HaveKeyWithValue("ltr_weekly_retention", "PT0S"),
					HaveKeyWithValue("ltr_monthly_retention", "PT0S"),
					HaveKeyWithValue("ltr_yearly_retention", "PT0S"),
					HaveKeyWithValue("ltr_week_of_year", BeNumerically("==", 1)),
				),
			)
		})

		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(mssqlDBServiceName, customMSSQLDBPlan["name"].(string), map[string]any{
				"cores":                     4,
				"max_storage_gb":            6,
				"db_name":                   "my-db-name",
				"server":                    "another-server",
				"sku_name":                  "GP_S",
				"short_term_retention_days": 2,
				"ltr_weekly_retention":      "P0W",
				"ltr_monthly_retention":     "P0M",
				"ltr_yearly_retention":      "P0Y",
				"ltr_week_of_year":          5,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("cores", BeNumerically("==", 4)),
					HaveKeyWithValue("max_storage_gb", BeNumerically("==", 6)),
					HaveKeyWithValue("db_name", "my-db-name"),
					HaveKeyWithValue("server", "another-server"),
					HaveKeyWithValue("sku_name", "GP_S"),
					HaveKeyWithValue("short_term_retention_days", BeNumerically("==", 2)),
					HaveKeyWithValue("ltr_weekly_retention", "P0W"),
					HaveKeyWithValue("ltr_monthly_retention", "P0M"),
					HaveKeyWithValue("ltr_yearly_retention", "P0Y"),
					HaveKeyWithValue("ltr_week_of_year", BeNumerically("==", 5)),
				),
			)
		})
	})
})
