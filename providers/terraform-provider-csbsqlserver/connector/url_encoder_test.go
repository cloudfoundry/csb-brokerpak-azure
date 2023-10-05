package connector_test

import (
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbsqlserver/connector"
)

var _ = Describe("Encoder", func() {
	server := "csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com"
	username := "fake_username"
	password := "fake_password"
	database := "db"
	port := 1433

	DescribeTable("change in encrypt property must generate a correct URL",
		func(encrypt string, wantedURL string) {
			got := connector.NewEncoder(server, username, password, database, encrypt, port).Encode()
			u, err := url.Parse(got)
			Expect(err).NotTo(HaveOccurred())
			uWanted, _ := url.Parse(wantedURL)
			Expect(u).To(Equal(uWanted))
		},
		Entry(
			"encrypt disable",
			"disable",
			"sqlserver://fake_username:fake_password@csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com:1433?database=db&encrypt=disable",
		),
		Entry(
			"encrypt true",
			"true",
			"sqlserver://fake_username:fake_password@csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com:1433?HostNameInCertificate=csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com&TrustServerCertificate=false&database=db&encrypt=true",
		),
		Entry(
			"encrypt different than true",
			"false",
			"sqlserver://fake_username:fake_password@csb-mssql-a74b4ec1-d534-4a7b-ac5e-3e644b7798b0.crvbjnvu3aun.us-west-2.rds.amazonaws.com:1433?database=db&encrypt=false",
		),
	)
})
