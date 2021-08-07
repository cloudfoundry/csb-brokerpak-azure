package postgresql_test

import (
	"acceptancetests/helpers"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL", func() {
	var serviceInstance helpers.ServiceInstance

	BeforeEach(func() {
		serviceInstance = helpers.CreateService("csb-azure-postgresql", "small")
	})

	AfterEach(func() {
		serviceInstance.Delete()
	})

	It("can be accessed by an app", func() {
		By("building the app")
		appDir := helpers.AppBuild("./postgresqlapp")
		defer os.RemoveAll(appDir)

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstartedBinaryBuildpack("postgresql", appDir)
		appTwo := helpers.AppPushUnstartedBinaryBuildpack("postgresql", appDir)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

		By("setting a key-value using the first app")
		key := helpers.RandomString()
		value := helpers.RandomString()
		appOne.PUT(value, key)

		By("getting the value using the first app")
		got := appTwo.GET(key)
		Expect(got).To(Equal(value))
	})
})
