package mssql_server_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Server Pair and Failover Group DB", func() {
	It("can be accessed by an app", func() {
		serversConfig := newDatabaseServerPair()

		By("Create CSB with server details")
		serviceBroker := helpers.CreateBroker(
			helpers.BrokerWithPrefix("csb-mssql-db"),
			helpers.BrokerWithEnv(helpers.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
		)
		defer serviceBroker.Delete()

		By("creating a primary server")
		serverInstancePrimary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", serviceBroker.Name, serversConfig.PrimaryConfig())
		defer serverInstancePrimary.Delete()

		// We have previously experienced problems with the CF CLI when doing things in parallel
		By("creating a secondary server in a different resource group")
		secondaryResourceGroupInstance := helpers.CreateServiceFromBroker("csb-azure-resource-group", "standard", serviceBroker.Name, serversConfig.SecondaryResourceGroupConfig())
		defer secondaryResourceGroupInstance.Delete()

		serverInstanceSecondary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", serviceBroker.Name, serversConfig.SecondaryConfig())
		defer serverInstanceSecondary.Delete()

		By("creating a database failover group on the server pair")
		fogName := random.Name(random.WithPrefix("fog"))
		dbFogInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db-failover-group", "small", serviceBroker.Name, map[string]string{
			"server_pair":   serversConfig.ServerPairTag,
			"instance_name": fogName,
		})
		defer dbFogInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.MSSQL)
		appTwo := helpers.AppPushUnstarted(apps.MSSQL)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := dbFogInstance.Bind(appOne)
		dbFogInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

		By("creating a schema using the first app")
		schema := random.Name(random.WithMaxLength(10))
		appOne.PUT("", schema)

		By("setting a key-value using the first app")
		keyOne := random.Hexadecimal()
		valueOne := random.Hexadecimal()
		appOne.PUT(valueOne, "%s/%s", schema, keyOne)

		By("getting the value using the second app")
		got := appTwo.GET("%s/%s", schema, keyOne)
		Expect(got).To(Equal(valueOne))

		By("triggering failover")
		failoverServiceInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-fog-run-failover", "standard", serviceBroker.Name, map[string]interface{}{
			"server_pair_name":  serversConfig.ServerPairTag,
			"server_pairs":      serversConfig.ServerPairsConfig(),
			"fog_instance_name": fogName,
		})
		defer failoverServiceInstance.Delete()

		By("setting another key-value")
		keyTwo := random.Hexadecimal()
		valueTwo := random.Hexadecimal()
		appTwo.PUT(valueTwo, "%s/%s", schema, keyTwo)

		By("getting the previously set values")
		Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))
		Expect(appTwo.GET("%s/%s", schema, keyTwo)).To(Equal(valueTwo))

		By("reverting the failover")
		failoverServiceInstance.Delete()

		By("dropping the schema")
		appOne.DELETE(schema)
	})
})

type DatabaseServerPair struct {
	ServerPairTag          string
	Username               string                   `json:"admin_username"`
	Password               string                   `json:"admin_password"`
	PrimaryServer          DatabaseServerPairMember `json:"primary"`
	SecondaryServer        DatabaseServerPairMember `json:"secondary"`
	SecondaryResourceGroup string                   `json:"-"`
}

type DatabaseServerPairMember struct {
	Name          string `json:"server_name"`
	ResourceGroup string `json:"resource_group"`
}

func newDatabaseServerPair() DatabaseServerPair {
	secondaryResourceGroup := random.Name(random.WithPrefix(metadata.ResourceGroup))
	return DatabaseServerPair{
		ServerPairTag: random.Name(random.WithMaxLength(10)),
		Username:      random.Name(random.WithMaxLength(10)),
		Password:      random.Password(),
		PrimaryServer: DatabaseServerPairMember{
			Name:          random.Name(random.WithPrefix("server")),
			ResourceGroup: metadata.ResourceGroup,
		},
		SecondaryServer: DatabaseServerPairMember{
			Name:          random.Name(random.WithPrefix("server")),
			ResourceGroup: secondaryResourceGroup,
		},
		SecondaryResourceGroup: secondaryResourceGroup,
	}
}

func (d DatabaseServerPair) PrimaryConfig() interface{} {
	return d.memberConfig(d.PrimaryServer.Name, "westus", d.PrimaryServer.ResourceGroup)
}

func (d DatabaseServerPair) SecondaryConfig() interface{} {
	return d.memberConfig(d.SecondaryServer.Name, "eastus", d.SecondaryServer.ResourceGroup)
}

func (d DatabaseServerPair) memberConfig(name, location, rg string) interface{} {
	return struct {
		Name          string `json:"instance_name"`
		Username      string `json:"admin_username"`
		Password      string `json:"admin_password"`
		Location      string `json:"location"`
		ResourceGroup string `json:"resource_group"`
	}{
		Name:          name,
		Username:      d.Username,
		Password:      d.Password,
		Location:      location,
		ResourceGroup: rg,
	}
}

func (d DatabaseServerPair) SecondaryResourceGroupConfig() interface{} {
	return struct {
		InstanceName string `json:"instance_name"`
		Location     string `json:"location"`
	}{
		InstanceName: d.SecondaryResourceGroup,
		Location:     "eastus",
	}
}

func (d DatabaseServerPair) ServerPairsConfig() interface{} {
	return map[string]interface{}{d.ServerPairTag: d}
}
