package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/mssql-serial/mssql_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group Existing", func() {
	It("can be accessed by an app", func() {
		By("creating a new resource group")
		rgConfig := resourceGroupConfig()
		resourceGroupInstance := helpers.CreateService("csb-azure-resource-group", "standard", rgConfig)
		defer resourceGroupInstance.Delete()

		By("creating primary and secondary DB servers in the resource group")
		serversConfig := newServerPair(rgConfig.Name)
		serverInstancePrimary := helpers.CreateService("csb-azure-mssql-server", "standard", serversConfig.PrimaryConfig())
		defer serverInstancePrimary.Delete()

		serverInstanceSecondary := helpers.CreateService("csb-azure-mssql-server", "standard", serversConfig.SecondaryConfig())
		defer serverInstanceSecondary.Delete()

		By("reconfiguring the CSB with DB server details")
		serversConfig.ReconfigureCSBWithServerDetails()

		By("creating a failover group service instance")
		fogConfig := failoverGroupConfig(serversConfig.ServerPairTag)
		initialFogInstance := helpers.CreateService("csb-azure-mssql-db-failover-group", "medium", fogConfig)
		defer initialFogInstance.Delete()

		By("pushing an unstarted app")
		app := helpers.AppPushUnstarted(apps.MSSQL)

		By("binding the app to the initial failover group service instance")
		bindingOne := initialFogInstance.Bind(app)

		By("starting the app")
		helpers.AppStart(app)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingOne.Credential()).To(helpers.HaveCredHubRef)

		By("creating a schema")
		schema := helpers.RandomShortName()
		app.PUT("", schema)

		By("setting a key-value")
		key := helpers.RandomHex()
		value := helpers.RandomHex()
		app.PUT(value, "%s/%s", schema, key)

		By("connecting to the existing failover group")
		dbFogInstance := helpers.CreateService("csb-azure-mssql-db-failover-group", "existing", fogConfig)
		defer dbFogInstance.Delete()

		By("purging the initial FOG instance")
		helpers.CF("purge-service-instance", "-f", initialFogInstance.Name())

		By("binding the app to the CSB service instance")
		bindingTwo := dbFogInstance.Bind(app)
		defer helpers.AppDelete(app) // app needs to be deleted before service instance

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingTwo.Credential()).To(helpers.HaveCredHubRef)

		By("getting the value set with the initial binding")
		got := app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the app")
		app.DELETE(schema)
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
	return  map[string]string{
		"instance_name": helpers.RandomName("fog"),
		"db_name":       helpers.RandomName("db"),
		"server_pair":   serverPairTag,
	}
}
