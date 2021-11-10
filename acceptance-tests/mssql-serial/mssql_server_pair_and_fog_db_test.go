package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/mssql-serial/mssql_helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Server Pair and Failover Group DB", func() {
	It("can be accessed by an app", func() {
		By("creating a primary server")
		serversConfig := newDatabaseServerPair()
		serverInstancePrimary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", helpers.DefaultBroker().Name, serversConfig.PrimaryConfig())
		defer serverInstancePrimary.Delete()

		// We have previously experienced problems with the CF CLI when doing things in parallel
		By("creating a secondary server in a different resource group")
		secondaryResourceGroupInstance := helpers.CreateServiceFromBroker("csb-azure-resource-group", "standard", helpers.DefaultBroker().Name, serversConfig.SecondaryResourceGroupConfig())
		defer secondaryResourceGroupInstance.Delete()
		serverInstanceSecondary := helpers.CreateServiceFromBroker("csb-azure-mssql-server", "standard", helpers.DefaultBroker().Name, serversConfig.SecondaryConfig())
		defer serverInstanceSecondary.Delete()

		By("reconfiguring the CSB with DB server details")
		serversConfig.ReconfigureCSBWithServerDetails()

		By("creating a database failover group on the server pair")
		fogName := helpers.RandomName("fog")
		dbFogInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-db-failover-group", "small", helpers.DefaultBroker().Name, map[string]string{
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
		schema := helpers.RandomShortName()
		appOne.PUT("", schema)

		By("setting a key-value using the first app")
		keyOne := helpers.RandomHex()
		valueOne := helpers.RandomHex()
		appOne.PUT(valueOne, "%s/%s", schema, keyOne)

		By("getting the value using the second app")
		got := appTwo.GET("%s/%s", schema, keyOne)
		Expect(got).To(Equal(valueOne))

		By("triggering failover")
		failoverServiceInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-fog-run-failover", "standard", helpers.DefaultBroker().Name, map[string]interface{}{
			"server_pair_name":  serversConfig.ServerPairTag,
			"server_pairs":      serversConfig.ServerPairsConfig(),
			"fog_instance_name": fogName,
		})
		defer failoverServiceInstance.Delete()

		By("setting another key-value")
		keyTwo := helpers.RandomHex()
		valueTwo := helpers.RandomHex()
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

func newDatabaseServerPair() mssql_helpers.DatabaseServerPair {
	secondaryResourceGroup := helpers.RandomName(metadata.ResourceGroup)
	return mssql_helpers.DatabaseServerPair{
		ServerPairTag: helpers.RandomShortName(),
		Username:      helpers.RandomShortName(),
		Password:      helpers.RandomPassword(),
		PrimaryServer: mssql_helpers.DatabaseServerPairMember{
			Name:          helpers.RandomName("server"),
			ResourceGroup: metadata.ResourceGroup,
		},
		SecondaryServer: mssql_helpers.DatabaseServerPairMember{
			Name:          helpers.RandomName("server"),
			ResourceGroup: secondaryResourceGroup,
		},
		SecondaryResourceGroup: secondaryResourceGroup,
	}
}
