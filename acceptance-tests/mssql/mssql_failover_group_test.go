package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group", func() {
	It("can be accessed by an app before and after failover", func() {
		By("creating a service instance")
		serviceInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-failover-group", "small-v2", helpers.DefaultBroker().Name)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.MSSQL)
		appTwo := helpers.AppPushUnstarted(apps.MSSQL)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the service instance")
		bindingOne := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingOne.Credential()).To(matchers.HaveCredHubRef)

		By("creating a schema using the first app")
		schema := random.Name(random.WithMaxLength(10))
		appOne.PUT("", schema)

		By("setting a key-value using the first app")
		keyOne := random.Hexadecimal()
		valueOne := random.Hexadecimal()
		appOne.PUT(valueOne, "%s/%s", schema, keyOne)

		By("getting the value using the second app")
		Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

		By("triggering failover")
		failoverServiceInstance := helpers.CreateServiceFromBroker("csb-azure-mssql-fog-run-failover", "standard", helpers.DefaultBroker().Name, failoverParameters(serviceInstance))
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
