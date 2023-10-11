package acceptance_test

import (
	"fmt"
	"regexp"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/matchers"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"csbbrokerpakazure/acceptance-tests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group", Label("mssql"), func() {
	It("can be accessed by an app before and after failover", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-azure-mssql-failover-group", "small-v2")
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.MSSQL))
		appTwo := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		bindingOne := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingOne.Credential()).To(matchers.HaveCredHubRef)

		By("creating a schema using the first app")
		schema := random.Name(random.WithMaxLength(10))
		appOne.PUT("", "%s?dbo=false", schema)

		By("setting a key-value using the first app")
		keyOne := random.Hexadecimal()
		valueOne := random.Hexadecimal()
		appOne.PUT(valueOne, "%s/%s", schema, keyOne)

		By("getting the value using the second app")
		Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

		By("triggering failover")
		failoverServiceInstance := services.CreateInstance(
			"csb-azure-mssql-fog-run-failover",
			"standard",
			services.WithParameters(failoverParameters(serviceInstance)),
		)
		defer failoverServiceInstance.Delete()

		By("setting another key-value")
		keyTwo := random.Hexadecimal()
		valueTwo := random.Hexadecimal()
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
