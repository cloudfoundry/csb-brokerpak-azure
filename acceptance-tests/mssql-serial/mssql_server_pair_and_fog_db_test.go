package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Server Pair and Failover Group DB", func() {
	It("can be accessed by an app", func() {
		By("creating a primary server")
		serversConfig := newDatabaseServerPair()
		serverInstancePrimary := helpers.CreateService("csb-azure-mssql-server", "standard", serversConfig.primaryConfig())
		defer serverInstancePrimary.Delete()

		// We have previously experienced problems with the CF CLI when doing things in parallel
		By("creating a secondary server")
		serverInstanceSecondary := helpers.CreateService("csb-azure-mssql-server", "standard", serversConfig.secondaryConfig())
		defer serverInstanceSecondary.Delete()

		By("reconfiguring the CSB with DB server details")
		serversConfig.reconfigureCSBWithServerDetails()

		By("creating a database failover group on the server pair")
		fogName := helpers.RandomName("fog")
		dbFogInstance := helpers.CreateService("csb-azure-mssql-db-failover-group", "small", map[string]string{
			"server_pair":   serversConfig.serverPairTag,
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
		failoverServiceInstance := helpers.CreateService("csb-azure-mssql-fog-run-failover", "standard", map[string]interface{}{
			"server_pair_name":  serversConfig.serverPairTag,
			"server_pairs":      serversConfig.serverPairsConfig(),
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

func newDatabaseServerPair() databaseServerPair {
	return databaseServerPair{
		serverPairTag: helpers.RandomShortName(),
		Username:      helpers.RandomShortName(),
		Password:      helpers.RandomPassword(),
		Primary: databaseServerPairMember{
			Name:          helpers.RandomName("server"),
			ResourceGroup: metadata.ResourceGroup,
		},
		Secondary: databaseServerPairMember{
			Name:          helpers.RandomName("server"),
			ResourceGroup: metadata.ResourceGroup,
		},
	}
}

type databaseServerPair struct {
	serverPairTag string
	Username      string                   `json:"admin_username"`
	Password      string                   `json:"admin_password"`
	Primary       databaseServerPairMember `json:"primary"`
	Secondary     databaseServerPairMember `json:"secondary"`
}

type databaseServerPairMember struct {
	Name          string `json:"server_name"`
	ResourceGroup string `json:"resource_group"`
}

func (d databaseServerPair) primaryConfig() interface{} {
	return d.memberConfig(d.Primary.Name, "westus")
}

func (d databaseServerPair) secondaryConfig() interface{} {
	return d.memberConfig(d.Secondary.Name, "eastus")
}

func (d databaseServerPair) memberConfig(name, location string) interface{} {
	return struct {
		Name     string `json:"instance_name"`
		Username string `json:"admin_username"`
		Password string `json:"admin_password"`
		Location string `json:"location"`
	}{
		Name:     name,
		Username: d.Username,
		Password: d.Password,
		Location: location,
	}
}

func (d databaseServerPair) serverPairsConfig() interface{} {
	return map[string]interface{}{d.serverPairTag: d}
}

func (d databaseServerPair) reconfigureCSBWithServerDetails() {
	helpers.SetBrokerEnv(
		helpers.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: d.serverPairsConfig()},
		helpers.EnvVar{Name: "GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS", Value: map[string]interface{}{"server_credential_pairs": d.serverPairsConfig()}},
	)
}
