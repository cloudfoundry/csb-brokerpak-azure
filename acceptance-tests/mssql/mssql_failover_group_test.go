package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL Failover Group", func() {
	It("can be accessed by an app before and after failover", func() {
		By("creating a service instance")
		serviceInstance := helpers.CreateService("csb-azure-mssql-failover-group", "small-v2")
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.MSSQL)
		appTwo := helpers.AppPushUnstarted(apps.MSSQL)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the service instance")
		bindingOne := serviceInstance.Bind(appOne)
		bindingTwo := serviceInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(bindingOne.Credential()).To(helpers.HaveCredHubRef)

		By("creating a schema using the first app")
		schema := helpers.RandomShortName()
		appOne.PUT("", schema)

		By("setting a key-value using the first app")
		keyOne := helpers.RandomString()
		valueOne := helpers.RandomString()
		appOne.PUT(valueOne, "%s/%s", schema, keyOne)

		By("getting the value using the second app")
		Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))

		By("triggering failover")
		bindingTwo.Unbind() // binding one owns the schema, so cannot be unbound
		failoverServiceInstance := helpers.CreateService("csb-azure-mssql-fog-run-failover", "standard", failoverParameters(serviceInstance))
		defer failoverServiceInstance.Delete()

		// Having to rebind is not ideal behaviour - see: https://www.pivotaltracker.com/story/show/179168006
		By("rebinding the second app")
		bindingTwo = serviceInstance.Bind(appTwo)
		helpers.AppRestage(appTwo)

		By("setting another key-value")
		keyTwo := helpers.RandomString()
		valueTwo := helpers.RandomString()
		appTwo.PUT(valueTwo, "%s/%s", schema, keyTwo)

		By("getting the previously set values")
		Expect(appTwo.GET("%s/%s", schema, keyOne)).To(Equal(valueOne))
		Expect(appTwo.GET("%s/%s", schema, keyTwo)).To(Equal(valueTwo))

		By("reverting the failover")
		bindingTwo.Unbind()
		failoverServiceInstance.Delete()

		By("dropping the schema")
		appOne.DELETE(schema)
	})
})
