package upgrade_test

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlDBFailoverGroupTest", Label("mssql-db-failover-group"), func() {
	When("upgrading broker version", Label("modern"), func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			rgConfig := resourceGroupConfig()
			serversConfig := newServerPair(rgConfig.Name)

			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-db-fo"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
			)
			defer serviceBroker.Delete()

			By("creating a new resource group")
			resourceGroupInstance := services.CreateInstance(
				"csb-azure-resource-group",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(rgConfig),
			)
			defer resourceGroupInstance.Delete()

			By("creating primary and secondary DB servers in the resource group")
			serverInstancePrimary := services.CreateInstance(
				"csb-azure-mssql-server",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(serversConfig.PrimaryConfig()),
			)
			defer serverInstancePrimary.Delete()

			serverInstanceSecondary := services.CreateInstance(
				"csb-azure-mssql-server",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(serversConfig.SecondaryConfig()),
			)
			defer serverInstanceSecondary.Delete()

			By("creating a failover group service instance")
			fogConfig := failoverGroupConfig(serversConfig.ServerPairTag)
			initialFogInstance := services.CreateInstance(
				"csb-azure-mssql-db-failover-group",
				"small",
				services.WithBroker(serviceBroker),
				services.WithParameters(fogConfig),
			)
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
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading previous services")
			resourceGroupInstance.Upgrade()
			serverInstancePrimary.Upgrade()
			serverInstanceSecondary.Upgrade()
			initialFogInstance.Upgrade()

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("updating the instance plan")
			initialFogInstance.Update("-p", "medium")

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("connecting to the existing failover group")
			dbFogInstance := services.CreateInstance(
				"csb-azure-mssql-db-failover-group",
				"existing",
				services.WithBroker(serviceBroker),
				services.WithParameters(fogConfig),
			)
			defer dbFogInstance.Delete()

			By("purging the initial FOG instance")
			initialFogInstance.Purge()

			By("creating new bindings and testing they still work")
			bindingOne := dbFogInstance.Bind(appOne)
			bindingTwo := dbFogInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)
			defer bindingOne.Unbind()
			defer bindingTwo.Unbind()

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

	When("upgrading broker version", Label("ancient"), func() {
		It("should continue to work", func() {
			By("pushing an ancient broker version")
			rgConfig := resourceGroupConfig()
			serversConfig := newServerPair(rgConfig.Name)

			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-db-fo"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithEnv(apps.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: serversConfig.ServerPairsConfig()}),
			)
			defer serviceBroker.Delete()

			By("creating a new resource group")
			resourceGroupInstance := services.CreateInstance(
				"csb-azure-resource-group",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(rgConfig),
			)
			defer resourceGroupInstance.Delete()

			By("creating primary and secondary DB servers in the resource group")
			serverInstancePrimary := services.CreateInstance(
				"csb-azure-mssql-server",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(serversConfig.PrimaryConfig()),
			)
			defer serverInstancePrimary.Delete()

			serverInstanceSecondary := services.CreateInstance(
				"csb-azure-mssql-server",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(serversConfig.SecondaryConfig()),
			)
			defer serverInstanceSecondary.Delete()

			By("creating a failover group service instance")
			fogConfig := failoverGroupAncientConfig(serversConfig.ServerPairTag, serversConfig.ServerPairsConfig())
			initialFogInstance := services.CreateInstance(
				"csb-azure-mssql-db-failover-group",
				"small",
				services.WithBroker(serviceBroker),
				services.WithParameters(fogConfig),
			)
			defer initialFogInstance.Delete()

			By("pushing the unstarted app")
			app := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(app)

			By("binding to the app")
			initialFogInstance.Bind(app)

			By("starting the app")
			apps.Start(app)

			By("creating a schema")
			schema := random.Name(random.WithMaxLength(10))
			app.PUT("", "%s?dbo=false", schema)

			By("setting a key-value")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
			app.PUT(valueOne, "%s/%s", schema, keyOne)

			By("getting the value")
			got := app.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading previous services")
			resourceGroupInstance.Upgrade()
			serverInstancePrimary.Upgrade()
			serverInstanceSecondary.Upgrade()
			initialFogInstance.Upgrade()

			By("getting the previously set value")
			got = app.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("updating the instance plan")
			initialFogInstance.Update("-p", "medium")

			By("getting the previously set value")
			got = app.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("connecting to the existing failover group")
			delete(fogConfig, "server_credential_pairs") // not accepted now that we have upgraded
			dbFogInstance := services.CreateInstance(
				"csb-azure-mssql-db-failover-group",
				"existing",
				services.WithBroker(serviceBroker),
				services.WithParameters(fogConfig),
			)
			defer dbFogInstance.Delete()

			By("purging the initial FOG instance")
			initialFogInstance.Purge()

			By("creating new bindings and testing they still work")
			binding := dbFogInstance.Bind(app)
			apps.Restage(app)
			defer binding.Unbind()

			By("getting the previously set value")
			Expect(app.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

			By("checking data can be written and read")
			keyTwo := random.Hexadecimal()
			valueTwo := random.Hexadecimal()
			app.PUT(valueTwo, "%s/%s", schema, keyTwo)

			got = app.GET("%s/%s", schema, keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("dropping the schema used to allow us to unbind")
			app.DELETE(schema)
		})
	})
})

func resourceGroupConfig() resourceConfig {
	return resourceConfig{
		Name:     random.Name(random.WithPrefix(metadata.ResourceGroup, "rg")),
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

func failoverGroupAncientConfig(serverPairTag string, serverCreds any) map[string]any {
	return map[string]any{
		"instance_name":           random.Name(random.WithPrefix("fog")),
		"db_name":                 random.Name(random.WithPrefix("db")),
		"server_pair":             serverPairTag,
		"server_credential_pairs": serverCreds,
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

func (d DatabaseServerPair) PrimaryConfig() any {
	return d.memberConfig(d.PrimaryServer.Name, "westus", d.PrimaryServer.ResourceGroup)
}

func (d DatabaseServerPair) SecondaryConfig() any {
	return d.memberConfig(d.SecondaryServer.Name, "eastus", d.SecondaryServer.ResourceGroup)
}

func (d DatabaseServerPair) memberConfig(name, location, rg string) any {
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

func (d DatabaseServerPair) SecondaryResourceGroupConfig() any {
	return struct {
		InstanceName string `json:"instance_name"`
		Location     string `json:"location"`
	}{
		InstanceName: d.SecondaryResourceGroup,
		Location:     "eastus",
	}
}

func (d DatabaseServerPair) ServerPairsConfig() any {
	return map[string]any{d.ServerPairTag: d}
}
