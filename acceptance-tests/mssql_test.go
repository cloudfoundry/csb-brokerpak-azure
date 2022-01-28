package acceptance_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL", Label("mssql"), func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-azure-mssql", "small-v2")
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.MSSQL))
		appTwo := apps.Push(apps.WithApp(apps.MSSQL))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

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
