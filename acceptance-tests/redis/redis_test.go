package redis_test

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/matchers"
	"acceptancetests/helpers/random"
	"acceptancetests/helpers/services"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Redis", func() {
	It("can be accessed by an app", func() {
		By("creating a service instance")
		serviceInstance := services.CreateInstance("csb-azure-redis", "small")
		defer serviceInstance.Delete()

		By("pushing the unstarted app twice")
		appOne := apps.Push(apps.WithApp(apps.Redis))
		appTwo := apps.Push(apps.WithApp(apps.Redis))
		defer apps.Delete(appOne, appTwo)

		By("binding the apps to the Redis service instance")
		binding := serviceInstance.Bind(appOne)
		serviceInstance.Bind(appTwo)

		By("starting the apps")
		apps.Start(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		Expect(binding.Credential()).To(matchers.HaveCredHubRef)

		By("setting a key-value using the first app")
		key := random.Hexadecimal()
		value := random.Hexadecimal()
		appOne.PUT(value, key)

		By("getting the value using the second app")
		got := appTwo.GET(key)
		Expect(got).To(Equal(value))
	})
})
