package csbsqlserver_test

import (
	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/csbsqlserver"
	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/testhelpers"
	"database/sql"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pborman/uuid"
)

var _ = Describe("csbsqlserver_binding resource", func() {
	var (
		port          int
		adminPassword string
		db            *sql.DB
	)

	BeforeEach(func() {
		adminPassword = testhelpers.RandomPassword()
		port = testhelpers.FreePort()
		session := testhelpers.StartServer(adminPassword, port)
		DeferCleanup(func() {
			session.Terminate().Wait(time.Minute)
		})
		db = testhelpers.Connect(testhelpers.AdminUser, adminPassword, testhelpers.TestDatabase, port)
	})

	It("can apply and destroy multiple bindings", func() {
		bindingUser1 := uuid.New()
		bindingPassword1 := testhelpers.RandomPassword()
		bindingUser2 := uuid.New()
		bindingPassword2 := testhelpers.RandomPassword()

		hcl := fmt.Sprintf(`
			provider "csbsqlserver" {
				server   = "%s"
				port     = "%d"
				database = "%s"
				username = "%s"
				password = "%s"
				encrypt  = "disable"
			}

			resource "csbsqlserver_binding" "binding1" {
				username = "%s"
				password = "%s"
				roles    = ["db_accessadmin", "db_datareader"]
			}

			resource "csbsqlserver_binding" "binding2" {
				username = "%s"
				password = "%s"
				roles    = ["db_accessadmin", "db_datareader"]
			}`,
			testhelpers.Server,
			port,
			testhelpers.TestDatabase,
			testhelpers.AdminUser,
			adminPassword,
			bindingUser1,
			bindingPassword1,
			bindingUser2,
			bindingPassword2,
		)

		resource.Test(GinkgoT(), resource.TestCase{
			IsUnitTest: true, // means we don't need to set TF_ACC
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"csbsqlserver": func() (*schema.Provider, error) { return csbsqlserver.Provider(), nil },
			},
			Steps: []resource.TestStep{{
				ResourceName: "csbsqlserver_binding",
				Config:       hcl,
				Check: func(state *terraform.State) error {
					for _, user := range []string{bindingUser1, bindingUser2} {
						if !testhelpers.UserExists(db, user) {
							return fmt.Errorf("user does not exist: %s", user)
						}
					}
					return nil
				},
			}},
			CheckDestroy: func(state *terraform.State) error {
				for _, user := range []string{bindingUser1, bindingUser2} {
					if testhelpers.UserExists(db, user) {
						return fmt.Errorf("user unexpectedly exists: %s", user)
					}
				}
				return nil
			},
		})
	})
})
