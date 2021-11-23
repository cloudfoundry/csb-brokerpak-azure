package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/mssql-serial/mssql_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlDBFailoverTest", func() {
	Context("When upgrading broker version", func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			brokerName := helpers.RandomName("mssql-db-fo")
			serviceBroker := helpers.PushAndStartBroker(brokerName, releasedBuildDir)
			defer serviceBroker.Delete()

			By("creating a new resource group")
			rgConfig := resourceGroupConfig()
			resourceGroupInstance := helpers.CreateServiceFromBroker("csb-azure-resource-group", "standard", brokerName, rgConfig)
			defer resourceGroupInstance.Delete()

			By("creating primary and secondary DB servers in the resource group")
			serversConfig := newServerPair(rgConfig.Name)
			serverInstancePrimary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", brokerName, serversConfig.PrimaryConfig())
			defer serverInstancePrimary.Delete()

			serverInstanceSecondary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", brokerName, serversConfig.SecondaryConfig())
			defer serverInstanceSecondary.Delete()

			By("reconfiguring the CSB with DB server details")
			serversConfig.ReconfigureCustomCSBWithServerDetails(brokerName)

			By("creating a failover group service instance")
			fogConfig := failoverGroupConfig(serversConfig.ServerPairTag)
			initialFogInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db-failover-group", "small", brokerName, fogConfig)
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
			schema := helpers.RandomShortName()
			appOne.PUT("", schema)

			By("setting a key-value using the first app")
			keyOne := helpers.RandomHex()
			valueOne := helpers.RandomHex()
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
			dbFogInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db-failover-group", "existing", brokerName, fogConfig)
			defer dbFogInstance.Delete()

			By("purging the initial FOG instance")
			helpers.CF("purge-service-instance", "-f", initialFogInstance.Name())

			By("creating new bindings and testing they still work")
			dbFogInstance.Bind(appOne)
			dbFogInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)
			defer dbFogInstance.Unbind(appOne)
			defer dbFogInstance.Unbind(appTwo)

			By("getting the previously set values")
			Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

			By("checking data can be written and read")
			keyTwo := helpers.RandomHex()
			valueTwo := helpers.RandomHex()
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
		Name:     helpers.RandomName("rg"),
		Location: "westus",
	}
}

type resourceConfig struct {
	Name     string `json:"instance_name"`
	Location string `json:"location"`
}

func newServerPair(resourceGroup string) mssql_helpers.DatabaseServerPair {
	return mssql_helpers.DatabaseServerPair{
		ServerPairTag: helpers.RandomShortName(),
		Username:      helpers.RandomShortName(),
		Password:      helpers.RandomPassword(),
		PrimaryServer: mssql_helpers.DatabaseServerPairMember{
			Name:          helpers.RandomName("server"),
			ResourceGroup: resourceGroup,
		},
		SecondaryServer: mssql_helpers.DatabaseServerPairMember{
			Name:          helpers.RandomName("server"),
			ResourceGroup: resourceGroup,
		},
	}
}

func failoverGroupConfig(serverPairTag string) map[string]string {
	return map[string]string{
		"instance_name": helpers.RandomName("fog"),
		"db_name":       helpers.RandomName("db"),
		"server_pair":   serverPairTag,
	}
}
