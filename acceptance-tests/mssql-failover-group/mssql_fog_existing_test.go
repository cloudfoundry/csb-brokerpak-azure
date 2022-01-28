package mssql_failover_group_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group Existing", func() {
	It("can be accessed by an app", func() {
		By("deploying the CSB")
		rgConfig := resourceGroupConfig()
		serversConfig := newServerPair(rgConfig.Name)
		serviceBroker := brokers.Create(
			brokers.WithPrefix("csb-mssql-db-fog"),
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
			"medium",
			services.WithBroker(serviceBroker),
			services.WithParameters(fogConfig),
		)
		defer initialFogInstance.Delete()

		By("pushing an unstarted app")
		app := apps.Push(apps.WithApp(apps.MSSQL))

		By("binding the app to the initial failover group service instance")
		bindingOne := initialFogInstance.Bind(app)

		By("starting the app")
		apps.Start(app)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingOne.Credential()).To(matchers.HaveCredHubRef)

		By("creating a schema")
		schema := random.Name(random.WithMaxLength(10))
		app.PUT("", schema)

		By("setting a key-value")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		app.PUT(value, "%s/%s", schema, key)

		By("connecting to the existing failover group")
		dbFogInstance := services.CreateInstance(
			"csb-azure-mssql-db-failover-group",
			"existing",
			services.WithBroker(serviceBroker),
			services.WithParameters(fogConfig),
		)
		defer dbFogInstance.Delete()

		By("purging the initial FOG instance")
		cf.Run("purge-service-instance", "-f", initialFogInstance.Name)

		By("binding the app to the CSB service instance")
		bindingTwo := dbFogInstance.Bind(app)
		defer apps.Delete(app) // app needs to be deleted before service instance

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingTwo.Credential()).To(matchers.HaveCredHubRef)

		By("getting the value set with the initial binding")
		got := app.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the app")
		app.DELETE(schema)
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
