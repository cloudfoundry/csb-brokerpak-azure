package csbsqlserver_test

import (
	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/csbsqlserver"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Provider Configuration", func() {
	var (
		server           string
		port             int
		providerUsername string
		providerPassword string
		database         string
		encrypt          string
		bindingUsername  string
		bindingPassword  string
		bindingRoles     []string
	)

	BeforeEach(func() {
		server = "flopsy.com"
		port = 1234
		providerUsername = "mopsy"
		providerPassword = "cottontail"
		database = "peter"
		encrypt = ""
		bindingUsername = "username"
		bindingPassword = "password"
		bindingRoles = make([]string, 0) // JSON of empty slice differs from nil slice
	})

	DescribeTable(
		"validation of parameters",
		func(cb func(), expectError string) {
			cb()

			roles, err := json.Marshal(bindingRoles)
			Expect(err).NotTo(HaveOccurred())

			hcl := fmt.Sprintf(`
				provider "csbsqlserver" {
					server   = "%s"
					port     = "%d"
					database = "%s"
					username = "%s"
					password = "%s"
					encrypt  = "%s"
				}

				resource "csbsqlserver_binding" "binding" {
					username = "%s"
					password = "%s"
					roles    = %s
				}`,
				server,
				port,
				database,
				providerUsername,
				providerPassword,
				encrypt,
				bindingUsername,
				bindingPassword,
				roles,
			)

			resource.Test(GinkgoT(), resource.TestCase{
				IsUnitTest: true, // means we don't need to set TF_ACC
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"csbsqlserver": func() (*schema.Provider, error) { return csbsqlserver.Provider(), nil },
				},
				Steps: []resource.TestStep{{
					ResourceName: "csbsqlserver",
					Config:       hcl,
					ExpectError:  regexp.MustCompile(expectError),
				}},
			})
		},
		Entry("server", func() { server = "not valid" }, `invalid URL value "not valid" for "server"`),
		Entry("port", func() { port = -1 }, `invalid port value -1 for "port", port values must be positive integers`),
		Entry("database", func() { database = "&&" }, `invalid value "&&" for identifier "database"`),
		Entry("provider username", func() { providerUsername = "&&" }, `invalid value "&&" for identifier "username"`),
		Entry("provider password", func() { providerPassword = "&&" }, `invalid password value for "password"`),
		Entry("encrypt", func() { encrypt = "maybe" }, `invalid value "maybe" for "encrypt"`),
		Entry("binding username", func() { bindingUsername = "&&" }, `invalid value "&&" for identifier "username"`),
		Entry("binding password", func() { bindingPassword = "&&" }, `invalid password value for "password"`),
		Entry("roles", func() { bindingRoles = []string{"&&"} }, `invalid value "&&" for element 0 of "roles"`),
	)
})
