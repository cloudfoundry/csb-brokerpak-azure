package mssql_test

import (
	"acceptancetests/helpers"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MSSQL", func() {
	var serviceInstanceName string

	BeforeEach(func() {
		serviceInstanceName = helpers.RandomName("mssql")
		helpers.CreateService("csb-azure-mssql", "small-v2", serviceInstanceName)
	})

	AfterEach(func() {
		helpers.DeleteService(serviceInstanceName)
	})

	It("can be accessed by an app", func() {
		By("building the app")
		appDir := helpers.AppBuild("./mssqlapp")
		defer os.RemoveAll(appDir)

		By("pushing the unstarted app twice")
		appOne := helpers.AppPushUnstartedBinaryBuildpack("mssql", appDir)
		appTwo := helpers.AppPushUnstartedBinaryBuildpack("mssql", appDir)
		defer helpers.AppDelete(appOne, appTwo)

		By("binding the apps to the service instance")
		bindingName := helpers.Bind(appOne, serviceInstanceName)
		helpers.Bind(appTwo, serviceInstanceName)

		By("starting the apps")
		helpers.AppStart(appOne, appTwo)

		By("checking that the app environment has a credhub reference for credentials")
		creds := helpers.GetBindingCredential(appOne, "csb-azure-mssql", bindingName)
		Expect(creds).To(HaveKey("credhub-ref"))

		appOneURL := fmt.Sprintf("http://%s.%s", appOne, helpers.DefaultSharedDomain())
		By("creating a schema using the first app")
		schema := helpers.RandomShortName()
		helpers.HTTPPut(fmt.Sprintf("%s/%s", appOneURL, schema), "")

		By("setting a key-value using the first app")
		key := helpers.RandomString()
		value := helpers.RandomString()
		helpers.HTTPPut(fmt.Sprintf("%s/%s/%s", appOneURL, schema, key), value)

		By("getting the value using the second app")
		got := helpers.HTTPGet(fmt.Sprintf("http://%s.%s/%s/%s", appTwo, helpers.DefaultSharedDomain(), schema, key))
		Expect(got).To(Equal(value))

		By("dropping the schema using the first app")
		helpers.HTTPDelete(fmt.Sprintf("%s/%s", appOneURL, schema))
	})
})
