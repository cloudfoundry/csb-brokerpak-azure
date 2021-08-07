package postgresql_test

import (
	"acceptancetests/apps"
	"acceptancetests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := helpers.CreateService("csb-azure-postgresql", "small")
		defer serviceInstance.Delete()

		By("pushing the unstarted app")
		app := helpers.AppPushUnstarted(apps.PostgeSQL)
		defer helpers.AppDelete(app)

		By("binding the app to the service instance")
		binding := serviceInstance.Bind(app)

		By("starting the app")
		helpers.AppStart(app)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(helpers.HaveCredHubRef)

		By("setting a key-value")
		key := helpers.RandomString()
		value := helpers.RandomString()
		app.PUT(value, key)

		By("getting the value")
		got := app.GET(key)
		Expect(got).To(Equal(value))
	})
})
