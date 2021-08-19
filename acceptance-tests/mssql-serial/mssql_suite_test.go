package mssql_test

import (
	"acceptancetests/helpers"
	"fmt"
	"os"
	"regexp"
	"testing"

	"code.cloudfoundry.org/jsonry"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMSSQL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MSSQL Serial Suite")
}

var metadata struct {
	ResourceGroup             string `jsonry:"name"`
	PreProvisionedSQLUsername string `jsonry:"masb_config.pre_provisioned_sql.username"`
	PreProvisionedSQLPassword string `jsonry:"masb_config.pre_provisioned_sql.password"`
	PreProvisionedSQLServer   string `jsonry:"masb_config.pre_provisioned_sql.server_name"`
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
})

func failoverParameters(instance helpers.ServiceInstance) interface{} {
	key := instance.CreateKey()
	defer key.Delete()

	var input struct {
		ServerName string `json:"sqlServerName"`
		Status     string `json:"status"`
	}
	key.Get(&input)

	resourceGroup := extractResourceGroup(input.Status)
	pairName := helpers.RandomName("server-pair")

	type failoverServer struct {
		Name          string `json:"server_name"`
		ResourceGroup string `json:"resource_group"`
	}

	type failoverServerPair struct {
		Primary   failoverServer `json:"primary"`
		Secondary failoverServer `json:"secondary"`
	}

	type failoverServerPairs map[string]failoverServerPair

	type output struct {
		FOGInstanceName string              `json:"fog_instance_name"`
		ServerPairName  string              `json:"server_pair_name"`
		ServerPairs     failoverServerPairs `json:"server_pairs"`
	}

	return output{
		FOGInstanceName: input.ServerName,
		ServerPairName:  pairName,
		ServerPairs: failoverServerPairs{
			pairName: failoverServerPair{
				Primary: failoverServer{
					Name:          fmt.Sprintf("%s-primary", input.ServerName),
					ResourceGroup: resourceGroup,
				},
				Secondary: failoverServer{
					Name:          fmt.Sprintf("%s-secondary", input.ServerName),
					ResourceGroup: resourceGroup,
				},
			},
		},
	}
}

func extractResourceGroup(status string) string {
	matches := regexp.MustCompile(`resourceGroups/(.+?)/`).FindStringSubmatch(status)
	Expect(matches).NotTo(BeNil())
	Expect(len(matches)).To(BeNumerically(">=", 2))
	return matches[1]
}
