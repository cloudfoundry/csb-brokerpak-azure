package upgrade_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"regexp"
)

var _ = Describe("UpgradeMssqlTest", func() {
	Context("When upgrading broker version", func(){
		It("should continue to work", func() {
			By("pushing latest released broker version")
			brokerName := helpers.RandomName("csb-mssql-failover-group")
			serviceBroker := helpers.PushAndStartBroker(brokerName, releasedBuildDir)
			defer serviceBroker.Delete()

			By("creating a service")
			serviceInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-failover-group", "small-v2", brokerName)
			defer serviceInstance.Delete()

			By("pushing the unstarted app twice")
			appOne := helpers.AppPushUnstarted(apps.MSSQL)
			appTwo := helpers.AppPushUnstarted(apps.MSSQL)
			defer helpers.AppDelete(appOne, appTwo)

			By("binding to the apps")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)

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

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)

			By("deleting bindings created before the upgrade")
			serviceInstance.Unbind(appOne)
			serviceInstance.Unbind(appTwo)

			By("creating new bindings and testing they still work")
			serviceInstance.Bind(appOne)
			serviceInstance.Bind(appTwo)
			helpers.AppRestage(appOne, appTwo)

			By("creating a schema using the first app")
			schema = helpers.RandomShortName()
			appOne.PUT("", schema)

			keyTwo := helpers.RandomHex()
			valueTwo := helpers.RandomHex()
			appOne.PUT(valueTwo, "%s/%s", schema, keyTwo)

			got = appTwo.GET("%s/%s", schema, keyTwo)
			Expect(got).To(Equal(valueTwo))

			By("updating the instance plan")
			serviceInstance.UpdateService("-p", "medium")

			By("triggering failover")
			failoverServiceInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-fog-run-failover", "standard", brokerName, failoverParameters(serviceInstance))
			defer failoverServiceInstance.Delete()

			By("getting the previously set values")
			Expect(appTwo.GET("%s/%s", schema, keyTwo)).To(Equal(valueTwo))

			By("checking data can still be written and read")
			keyThree := helpers.RandomHex()
			valueThree := helpers.RandomHex()
			appOne.PUT(valueThree, "%s/%s", schema, keyThree)

			got = appTwo.GET("%s/%s", schema, keyThree)
			Expect(got).To(Equal(valueThree))

			By("dropping the schema used to allow us to unbind")
			appOne.DELETE(schema)
		})
	})
})

func failoverParameters(instance helpers.ServiceInstance) interface{} {
	key := instance.CreateKey()
	defer key.Delete()

	var input struct {
		ServerName string `json:"sqlServerName"`
		Status     string `json:"status"`
	}
	key.Get(&input)

	resourceGroup := extractResourceGroup(input.Status)
	pairName := helpers.RandomName("server-pair")

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