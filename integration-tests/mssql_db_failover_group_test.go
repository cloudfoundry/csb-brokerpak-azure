package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	mssqlDBFailoverGroupServiceName             = "csb-azure-mssql-db-failover-group"
	mssqlDBFailoverGroupServiceID               = "d7ba0e8e-4480-4543-a504-b57e1dd1f1ad"
	mssqlDBFailoverGroupServiceDisplayName      = "Azure SQL Failover Group on Existing Server Pairs"
	mssqlDBFailoverGroupServiceDescription      = "Manages auto failover group db's on existing Azure SQL server pairs"
	mssqlDBFailoverGroupServiceDocumentationURL = "https://docs.microsoft.com/en-us/azure/sql-database/sql-database-auto-failover-group/"
	mssqlDBFailoverGroupServiceSupportURL       = "https://docs.microsoft.com/en-us/azure/sql-database/sql-database-auto-failover-group/"
)

var _ = Describe("MSSQL DB Auto-failover group", Label("MSSQL Auto-failover group"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("should publish in the catalog", func() {
		catalog, err := broker.Catalog()
		Expect(err).NotTo(HaveOccurred())

		service := testframework.FindService(catalog, mssqlDBFailoverGroupServiceName)
		Expect(service.ID).To(Equal(mssqlDBFailoverGroupServiceID))
		Expect(service.Description).To(Equal(mssqlDBFailoverGroupServiceDescription))
		Expect(service.Tags).To(ConsistOf("azure", "mssql", "sqlserver", "dr", "failover", "preview"))
		Expect(service.Metadata.ImageUrl).To(ContainSubstring("data:image/png;base64,"))
		Expect(service.Metadata.DisplayName).To(Equal(mssqlDBFailoverGroupServiceDisplayName))
		Expect(service.Metadata.DocumentationUrl).To(Equal(mssqlDBFailoverGroupServiceDocumentationURL))
		Expect(service.Metadata.SupportUrl).To(Equal(mssqlDBFailoverGroupServiceSupportURL))
		Expect(service.Plans).To(
			ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("small"),
					ID:   Equal("35a7e882-9e27-4e5a-a292-9c3f3da10873"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("medium"),
					ID:   Equal("b653c8a6-4094-4103-8958-4630a42e1c49"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("large"),
					ID:   Equal("609b668e-0cf8-4512-9f42-ef684c0c8d8d"),
				}),
				MatchFields(IgnoreExtras, Fields{
					Name: Equal("existing"),
					ID:   Equal("669661c1-7fe6-4c59-8004-63905e79a508"),
				}),
			),
		)
	})

	Describe("provisioning", func() {
		DescribeTable("should check property constraints",
			func(params map[string]any, expectedErrorMsg string) {
				params["server_pair"] = "preprovisioned-server-name"
				_, err := broker.Provision(mssqlDBFailoverGroupServiceName, "small", params)
				Expect(err).To(MatchError(ContainSubstring(expectedErrorMsg)))
			},
			Entry(
				"invalid cores",
				map[string]any{"cores": 0},
				"cores: Must be greater than or equal to 1",
			),
			Entry(
				"invalid cores",
				map[string]any{"cores": 3},
				"cores: Must be a multiple of 2",
			),
			Entry(
				"invalid cores",
				map[string]any{"cores": 82},
				"cores: Must be less than or equal to 80",
			),
			Entry(
				"invalid instance_name",
				map[string]any{"instance_name": stringOfLen(65)},
				"instance_name: String length must be less than or equal to 63",
			),
			Entry(
				"invalid instance_name",
				map[string]any{"instance_name": "short"},
				"instance_name: String length must be greater than or equal to 6",
			),
			Entry(
				"invalid db_name",
				map[string]any{"db_name": stringOfLen(65)},
				"db_name: String length must be less than or equal to 64",
			),
			Entry(
				"invalid read_write_endpoint_failover_policy",
				map[string]any{"read_write_endpoint_failover_policy": "something-fishy"},
				"read_write_endpoint_failover_policy: read_write_endpoint_failover_policy must be one of the following:",
			),
			Entry(
				"invalid short_term_retention_days",
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
			instanceID, err := broker.Provision(mssqlDBFailoverGroupServiceName, "small", map[string]any{
				"server_pair": "preprovisioned-server-name",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("cores", BeNumerically("==", 2)),
					HaveKeyWithValue("max_storage_gb", BeNumerically("==", 5)),
					HaveKeyWithValue("db_name", fmt.Sprintf("csb-fog-db-%s", instanceID)),
					HaveKeyWithValue("sku_name", BeEmpty()),
					HaveKeyWithValue("short_term_retention_days", BeNumerically("==", 7)),
					HaveKeyWithValue("ltr_weekly_retention", "PT0S"),
					HaveKeyWithValue("ltr_monthly_retention", "PT0S"),
					HaveKeyWithValue("ltr_yearly_retention", "PT0S"),
					HaveKeyWithValue("ltr_week_of_year", BeNumerically("==", 1)),
					HaveKeyWithValue("read_write_endpoint_failover_policy", "Automatic"),
					HaveKeyWithValue("skip_provider_registration", false),
					HaveKeyWithValue("existing", false),
				),
			)
		})
		It("should allow properties to be set on provision", func() {
			_, err := broker.Provision(mssqlDBFailoverGroupServiceName, "small", map[string]any{
				"cores":                               4,
				"max_storage_gb":                      6,
				"db_name":                             "my-db-name",
				"server_pair":                         "another-server",
				"sku_name":                            "GP_S",
				"short_term_retention_days":           2,
				"ltr_weekly_retention":                "P1W",
				"ltr_monthly_retention":               "P2M",
				"ltr_yearly_retention":                "P5Y",
				"ltr_week_of_year":                    5,
				"read_write_endpoint_failover_policy": "Manual",
				"skip_provider_registration":          true,
				"existing":                            true,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(mockTerraform.FirstTerraformInvocationVars()).To(
				SatisfyAll(
					HaveKeyWithValue("cores", BeNumerically("==", 4)),
					HaveKeyWithValue("max_storage_gb", BeNumerically("==", 6)),
					HaveKeyWithValue("db_name", "my-db-name"),
					HaveKeyWithValue("server_pair", "another-server"),
					HaveKeyWithValue("sku_name", "GP_S"),
					HaveKeyWithValue("short_term_retention_days", BeNumerically("==", 2)),
					HaveKeyWithValue("ltr_weekly_retention", "P1W"),
					HaveKeyWithValue("ltr_monthly_retention", "P2M"),
					HaveKeyWithValue("ltr_yearly_retention", "P5Y"),
					HaveKeyWithValue("ltr_week_of_year", BeNumerically("==", 5)),
					HaveKeyWithValue("read_write_endpoint_failover_policy", "Manual"),
					HaveKeyWithValue("skip_provider_registration", true),
					HaveKeyWithValue("existing", true),
				),
			)
		})
	})
})
