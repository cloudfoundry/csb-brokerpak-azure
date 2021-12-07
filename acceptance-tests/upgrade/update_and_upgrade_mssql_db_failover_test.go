package upgrade_test

import (
	"acceptancetests/helpers"
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlDBFailoverTest", func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			rgConfig := resourceGroupConfig()
			serversConfig := newServerPair(rgConfig.Name)

			serviceBroker := helpers.CreateBroker(
				helpers.BrokerWithPrefix("csb-db-fo"),
				helpers.BrokerFromDir(releasedBuildDir),
				helpers.BrokerWithEnv(
					apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()},
				))

			defer serviceBroker.Delete()

			By("creating a new resource group")
			resourceGroupInstance := helpers.CreateServiceFromBroker("csb-azure-resource-group", "standard", serviceBroker.Name, rgConfig)
			defer resourceGroupInstance.Delete()

			By("creating primary and secondary DB servers in the resource group")
			serverInstancePrimary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", serviceBroker.Name, serversConfig.PrimaryConfig())
			defer serverInstancePrimary.Delete()

			serverInstanceSecondary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", serviceBroker.Name, serversConfig.SecondaryConfig())
			defer serverInstanceSecondary.Delete()

			By("creating a failover group service instance")
			fogConfig := failoverGroupConfig(serversConfig.ServerPairTag)
			initialFogInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db-failover-group", "small", serviceBroker.Name, fogConfig)
			defer initialFogInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MSSQL))
			appTwo := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			initialFogInstance.Bind(appOne)
			initialFogInstance.Bind(appTwo)

			By("starting the apps")
			apps.Start(appOne, appTwo)

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

			By("pushing the development version of the broker")
			serviceBroker.Update(developmentBuildDir)

			By("updating the instance plan")
			initialFogInstance.UpdateService("-p", "medium")

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("connecting to the existing failover group")
			dbFogInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db-failover-group", "existing", serviceBroker.Name, fogConfig)
			defer dbFogInstance.Delete()

			By("purging the initial FOG instance")
			cf.Run("purge-service-instance", "-f", initialFogInstance.Name())

			By("creating new bindings and testing they still work")
			dbFogInstance.Bind(appOne)
			dbFogInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)
			defer dbFogInstance.Unbind(appOne)
			defer dbFogInstance.Unbind(appTwo)

			By("getting the previously set values")
			Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

			By("checking data can be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			appOne.PUT(valueTwo, "%s/%s", schema, keyTwo)

			got = appTwo.GET("%s/%s", schema, keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)
		})
	})
})

func resourceGroupConfig() resourceConfig {
	return resourceConfig{
		Name:     random.Name(random.WithPrefix("rg")),
		Location: "westus",
	}
}

type resourceConfig struct {
	Name     string `json:"instance_name"`
	Location string `json:"location"`
}

func newServerPair(resourceGroup string) DatabaseServerPair {
	return DatabaseServerPair{
		ServerPairTag: random.Name(random.WithMaxLength(10)),
		Username:      random.Name(random.WithMaxLength(10)),
		Password:      random.Password(),
		PrimaryServer: DatabaseServerPairMember{
			Name:          random.Name(random.WithPrefix("server")),
			ResourceGroup: resourceGroup,
		},
		SecondaryServer: DatabaseServerPairMember{
			Name:          random.Name(random.WithPrefix("server")),
			ResourceGroup: resourceGroup,
		},
	}
}

func failoverGroupConfig(serverPairTag string) map[string]string {
	return map[string]string{
		"instance_name": random.Name(random.WithPrefix("fog")),
		"db_name":       random.Name(random.WithPrefix("db")),
		"server_pair":   serverPairTag,
	}
}

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
