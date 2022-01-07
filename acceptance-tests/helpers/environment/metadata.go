package environment

import (
	"os"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/gomega"
)

type Metadata struct {
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

func ReadMetadata() Metadata {
	file := os.Getenv("ENVIRONMENT_LOCK_METADATA")
	Expect(file).NotTo(BeEmpty(), "You must set the ENVIRONMENT_LOCK_METADATA environment variable")

	contents, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())

	var metadata Metadata
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
	return metadata
}
