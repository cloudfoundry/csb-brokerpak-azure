package mssql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"
	"acceptancetests/helpers/random"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := helpers.CreateServiceFromBroker("csb-azure-mssql", "small-v2", helpers.DefaultBroker().Name)
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstarted(apps.MSSQL)
		appTwo := helpers.AppPushUnstarted(apps.MSSQL)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

		By("creating a schema using the first app")
		schema := random.Name(random.WithMaxLength(10))
		appOne.PUT("", schema)

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appOne.PUT(value, "%s/%s", schema, key)

		By("getting the value using the second app")
		got := appTwo.GET("%s/%s", schema, key)
		Expect(got).To(Equal(value))

		By("dropping the schema using the first app")
		appOne.DELETE(schema)
	})
})
