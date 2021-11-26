package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
	"acceptancetests/mssql-serial/mssql_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlDBFailoverTest", func() {
	When("upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := helpers.CreateBroker(
				helpers.BrokerWithPrefix("csb-db-fo"),
				helpers.BrokerFromDir(releasedBuildDir),
			)

			defer serviceBroker.Delete()

			By("creating a new resource group")
			rgConfig := resourceGroupConfig()
			resourceGroupInstance := helpers.CreateServiceFromBroker("csb-azure-resource-group", "standard", serviceBroker.Name, rgConfig)
			defer resourceGroupInstance.Delete()

			By("creating primary and secondary DB servers in the resource group")
			serversConfig := newServerPair(rgConfig.Name)
			serverInstancePrimary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", serviceBroker.Name, serversConfig.PrimaryConfig())
			defer serverInstancePrimary.Delete()

			serverInstanceSecondary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", serviceBroker.Name, serversConfig.SecondaryConfig())
			defer serverInstanceSecondary.Delete()

			By("reconfiguring the CSB with DB server details")
			serversConfig.ReconfigureCustomCSBWithServerDetails(serviceBroker.Name)

			By("creating a failover group service instance")
			fogConfig := failoverGroupConfig(serversConfig.ServerPairTag)
			initialFogInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db-failover-group", "small", serviceBroker.Name, fogConfig)
			defer initialFogInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.MSSQL)
			appTwo := helpers.AppPushUnstarted(apps.MSSQL)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding to the apps")
			initialFogInstance.Bind(appOne)
			initialFogInstance.Bind(appTwo)

			By("starting the apps")
			helpers.AppStart(appOne, appTwo)

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
			helpers.AppRestage(appOne, appTwo)
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

func newServerPair(resourceGroup string) mssql_helpers.DatabaseServerPair {
	return mssql_helpers.DatabaseServerPair{
		ServerPairTag: random.Name(random.WithMaxLength(10)),
		Username:      random.Name(random.WithMaxLength(10)),
		Password:      random.Password(),
		PrimaryServer: mssql_helpers.DatabaseServerPairMember{
			Name:          random.Name(random.WithPrefix("server")),
			ResourceGroup: resourceGroup,
		},
		SecondaryServer: mssql_helpers.DatabaseServerPairMember{
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
