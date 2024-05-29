package upgrade_test

import (
	"fmt"
	"regexp"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/brokers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpgradeMssqlFailoverGroupTest", Label("mssql-failover-group"), func() {
	When("upgrading broker version", Label("modern"), func() {
		It("should continue to work", func() {
			By("pushing latest released broker version")
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-mssql-fog"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-azure-mssql-failover-group",
				"small-v2",
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := apps.Push(apps.WithApp(apps.MSSQL))
			appTwo := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appOne, appTwo)

			By("binding to the apps")
			bindingOne := serviceInstance.Bind(appOne)
			bindingTwo := serviceInstance.Bind(appTwo)

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

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("updating the instance plan")
			serviceInstance.Update("-p", "medium")

			By("getting the previously set value using the second app")
			got = appTwo.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("triggering failover")
			failoverServiceInstance := services.CreateInstance(
				"csb-azure-mssql-fog-run-failover",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(failoverParameters(serviceInstance)),
			)
			defer failoverServiceInstance.Delete()

			By("getting the previously set values")
			Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)

			By("deleting bindings created before the upgrade")
			bindingOne.Unbind()
			bindingTwo.Unbind()

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			apps.Restage(appOne, appTwo)

			By("creating a schema using the first app")
			schema = random.Name(random.WithMaxLength(10))
			appOne.PUT("", schema)

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
			serviceBroker := brokers.Create(
				brokers.WithPrefix("csb-mssql-fog"),
				brokers.WithSourceDir(releasedBuildDir),
				brokers.WithReleaseEnv(releasedBuildDir),
			)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := services.CreateInstance(
				"csb-azure-mssql-failover-group",
				"small-v2",
				services.WithBroker(serviceBroker),
			)
			defer serviceInstance.Delete()

			By("pushing the unstarted app")
			appOne := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appOne)

			By("binding to the app")
			serviceInstance.Bind(appOne)

			By("starting the app")
			apps.Start(appOne)

			By("creating a schema")
			schema := random.Name(random.WithMaxLength(10))
			appOne.PUT("", "%s?dbo=false", schema)

			By("setting a key-value")
			keyOne := random.Hexadecimal()
			valueOne := random.Hexadecimal()
			appOne.PUT(valueOne, "%s/%s", schema, keyOne)

			By("getting the value")
			got := appOne.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("pushing the development version of the broker")
			serviceBroker.UpgradeBroker(developmentBuildDir)

			By("upgrading service instance")
			serviceInstance.Upgrade()

			By("getting the previously set value using the second app")
			got = appOne.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("updating the instance plan")
			serviceInstance.Update("-p", "medium")

			By("getting the previously set value using the second app")
			got = appOne.GET("%s/%s", schema, keyOne)
			Expect(got).To(Equal(valueOne))

			By("triggering failover")
			failoverServiceInstance := services.CreateInstance(
				"csb-azure-mssql-fog-run-failover",
				"standard",
				services.WithBroker(serviceBroker),
				services.WithParameters(failoverParameters(serviceInstance)),
			)
			defer failoverServiceInstance.Delete()

			// Because of https://github.com/cloudfoundry/csb-brokerpak-azure/commit/40c6fd9744fe696296c84e3783b5f20af07f9ad9
			// the binding users created before the upgrade will not exist in the replica
			By("pushing and binding another app")
			appTwo := apps.Push(apps.WithApp(apps.MSSQL))
			defer apps.Delete(appTwo)
			serviceInstance.Bind(appTwo)
			apps.Start(appTwo)

			By("getting the previously set values")
			Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

			apps.Delete(appTwo)
			failoverServiceInstance.Delete()
			appOne.DELETE(schema)
		})
	})
})

func failoverParameters(instance *services.ServiceInstance) any {
	key := instance.CreateServiceKey()
	defer key.Delete()

	var input struct {
		ServerName string `json:"sqlServerName"`
		Status     string `json:"status"`
	}
	key.Get(&input)

	resourceGroup := extractResourceGroup(input.Status)
	pairName := random.Name(random.WithPrefix("server-pair"))

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
