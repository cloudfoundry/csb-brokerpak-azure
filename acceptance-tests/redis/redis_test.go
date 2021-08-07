package redis_test

import (
	"acceptancetests/helpers"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Redis", func() {
	var serviceInstance helpers.ServiceInstance

	BeforeEach(func() {
		serviceInstance = helpers.CreateService("csb-azure-redis", "small")
	})

	AfterEach(func() {
		serviceInstance.Delete()
	})

	It("can be accessed by an app", func() {
		By("building the app")
		appDir := helpers.AppBuild("./redisapp")
		defer os.RemoveAll(appDir)

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstartedBinaryBuildpack("redis", appDir)
		appTwo := helpers.AppPushUnstartedBinaryBuildpack("redis", appDir)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the Redis service instance")
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

		By("getting the value using the second app")
		got := appTwo.GET(key)
		Expect(got).To(Equal(value))
	})
})
