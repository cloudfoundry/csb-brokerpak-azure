package mssql_failover_group_test

import (
	"os"
	"testing"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMssqlFailoverGroup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MssqlFailoverGroup Suite")
}

var metadata struct {
	ResourceGroup             string `jsonry:"name"`
	PreProvisionedSQLUsername string `jsonry:"masb_config.pre_provisioned_sql.username"`
	PreProvisionedSQLPassword string `jsonry:"masb_config.pre_provisioned_sql.password"`
	PreProvisionedSQLServer   string `jsonry:"masb_config.pre_provisioned_sql.server_name"`
	PreProvisionedSQLLocation string `jsonry:"masb_config.location"`
	PreProvisionedFOGUsername string `jsonry:"masb_config.pre_provisioned_fog_sql.username"`
	PreProvisionedFOGPassword string `jsonry:"masb_config.pre_provisioned_fog_sql.password"`
	PreProvisionedFOGServer   string `jsonry:"masb_config.pre_provisioned_fog_sql.server_name"`
	PreProvisionedFOGLocation string `jsonry:"masb_config.pre_provisioned_fog_sql.location"`
}

var _ = BeforeSuite(func() {
	file := os.Getenv("ENVIRONMENT_LOCK_METADATA")
	Expect(file).NotTo(BeEmpty(), "You must set the ENVIRONMENT_LOCK_METADATA environment variable")

	contents, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())

	Expect(jsonry.Unmarshal(contents, &metadata)).NotTo(HaveOccurred())
	Expect(metadata.ResourceGroup).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedSQLUsername).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedSQLPassword).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedSQLServer).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedSQLLocation).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedFOGUsername).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedFOGPassword).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedFOGServer).NotTo(BeEmpty())
	Expect(metadata.PreProvisionedFOGLocation).NotTo(BeEmpty())
})
