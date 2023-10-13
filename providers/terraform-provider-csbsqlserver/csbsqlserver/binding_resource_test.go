package csbsqlserver_test

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"

	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbsqlserver/csbsqlserver"
	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbsqlserver/testhelpers"
)

const (
	providerName = "csbsqlserver"
)

var _ = Describe("csbsqlserver_binding resource", func() {

	Context("database exists", func() {
		When("bindings are created", func() {

			It("can apply and destroy multiple bindings", func() {

				var (
					adminPassword = testhelpers.RandomPassword()
					port          = testhelpers.FreePort()
				)

				shutdownServerFn := testhelpers.StartServer(adminPassword, port)
				DeferCleanup(func() { shutdownServerFn(time.Minute) })

				cnf := createTestCaseCnf(adminPassword, port)

				resource.Test(GinkgoT(), getTestCase(cnf, getMandatoryStep(cnf)))
			})
		})
	})

	Context("database does not exists", func() {
		When("binding is created", func() {
			It("should create a database", func() {
				var (
					adminPassword = testhelpers.RandomPassword()
					port          = testhelpers.FreePort()
				)

				shutdownServerFn := testhelpers.StartServer(adminPassword, port, testhelpers.WithSPConfigure())
				DeferCleanup(func() { shutdownServerFn(time.Minute) })

				cnf := createTestCaseCnf(adminPassword, port)

				resource.Test(GinkgoT(), getTestCase(cnf, getMandatoryStep(cnf)))
			})
		})
	})
})

type testCaseCnf struct {
	ResourceBindingOneName string
	ResourceBindingTwoName string
	BindingUserOne         string
	BindingUserTwo         string
	BindingPasswordOne     string
	BindingPasswordTwo     string
	DatabaseName           string
	AdminPassword          string
	Port                   int
	provider               *schema.Provider
	ExpectError            *regexp.Regexp
}

func createTestCaseCnf(adminPassword string, port int) testCaseCnf {
	return testCaseCnf{
		ResourceBindingOneName: fmt.Sprintf("%s.binding1", csbsqlserver.ResourceNameKey),
		ResourceBindingTwoName: fmt.Sprintf("%s.binding2", csbsqlserver.ResourceNameKey),
		BindingUserOne:         fmt.Sprintf("user_one_%s", uuid.NewString()),
		BindingUserTwo:         fmt.Sprintf("user_two_%s", uuid.NewString()),
		BindingPasswordOne:     testhelpers.RandomPassword(),
		BindingPasswordTwo:     testhelpers.RandomPassword(),
		DatabaseName:           testhelpers.RandomDatabaseName(),
		AdminPassword:          adminPassword,
		Port:                   port,
		provider:               initTestProvider(),
	}
}

func getTestCase(cnf testCaseCnf, steps ...resource.TestStep) resource.TestCase {
	var (
		bindingUser1, bindingUser2 = cnf.BindingUserOne, cnf.BindingUserTwo
		databaseName               = cnf.DatabaseName
		provider                   = cnf.provider
		db                         = testhelpers.Connect(testhelpers.AdminUser, cnf.AdminPassword, databaseName, cnf.Port)
	)

	return resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: getTestProviderFactories(provider),
		Steps:             steps,
		CheckDestroy: func(state *terraform.State) error {
			for _, user := range []string{bindingUser1, bindingUser2} {
				if testhelpers.UserExists(db, user) {
					return fmt.Errorf("user unexpectedly exists: %s", user)
				}
			}
			return nil
		},
	}
}

func getMandatoryStep(cnf testCaseCnf, extraTestCheckFunc ...resource.TestCheckFunc) resource.TestStep {
	var (
		tfStateResourceBinding1Name        = cnf.ResourceBindingOneName
		tfStateResourceBinding2Name        = cnf.ResourceBindingTwoName
		bindingUser1, bindingUser2         = cnf.BindingUserOne, cnf.BindingUserTwo
		bindingPassword1, bindingPassword2 = cnf.BindingPasswordOne, cnf.BindingPasswordTwo
		databaseName                       = cnf.DatabaseName
		db                                 = testhelpers.Connect(testhelpers.AdminUser, cnf.AdminPassword, databaseName, cnf.Port)
	)

	return resource.TestStep{
		ResourceName: csbsqlserver.ResourceNameKey,
		Config:       testGetConfiguration(cnf.Port, cnf.AdminPassword, bindingUser1, bindingPassword1, bindingUser2, bindingPassword2, databaseName),
		ExpectError:  cnf.ExpectError,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "username", bindingUser1),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "password", bindingPassword1),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.0", "db_ddladmin"),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.1", "db_datareader"),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.2", "db_datawriter"),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.3", "db_accessadmin"),
			resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "username", bindingUser2),
			resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "password", bindingPassword2),
			resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "roles.0", "db_ddladmin"),
			resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "roles.1", "db_datareader"),
			resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "roles.2", "db_datawriter"),
			resource.TestCheckResourceAttr(tfStateResourceBinding2Name, "roles.3", "db_accessadmin"),
			testCheckDatabaseExists(db, databaseName),
			testCheckUserExists(db, bindingUser1),
			testCheckUserExists(db, bindingUser2),
			func(state *terraform.State) error {
				for _, checkFn := range extraTestCheckFunc {
					if err := checkFn(state); err != nil {
						return err
					}
				}
				return nil
			},
		),
	}
}

func getStepOnlyBindingOne(cnf testCaseCnf, extraTestCheckFunc ...resource.TestCheckFunc) resource.TestStep {
	var (
		tfStateResourceBinding1Name = cnf.ResourceBindingOneName
		bindingUser1                = cnf.BindingUserOne
		bindingPassword1            = cnf.BindingPasswordOne
		databaseName                = cnf.DatabaseName
		db                          = testhelpers.Connect(testhelpers.AdminUser, cnf.AdminPassword, databaseName, cnf.Port)
	)

	return resource.TestStep{
		ResourceName: csbsqlserver.ResourceNameKey,
		Config:       testGetConfigurationOnlyBindingOne(cnf.Port, cnf.AdminPassword, bindingUser1, bindingPassword1, databaseName),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "username", bindingUser1),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "password", bindingPassword1),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.0", "db_ddladmin"),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.1", "db_datareader"),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.2", "db_datawriter"),
			resource.TestCheckResourceAttr(tfStateResourceBinding1Name, "roles.3", "db_accessadmin"),
			testCheckDatabaseExists(db, databaseName),
			testCheckUserExists(db, bindingUser1),
			testCheckUserDoesNotExists(db, cnf.BindingUserTwo),
			resource.ComposeAggregateTestCheckFunc(extraTestCheckFunc...),
		),
	}
}

func testCheckUserExists(db *sql.DB, username string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		if !testhelpers.UserExists(db, username) {
			return fmt.Errorf("user does not exist: %s", username)
		}
		return nil
	}
}

func testCheckUserDoesNotExists(db *sql.DB, username string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		if testhelpers.UserExists(db, username) {
			return fmt.Errorf("the user must not exist: %s", username)
		}
		return nil
	}
}

func testCheckDatabaseExists(db *sql.DB, databaseName string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		statement := `SELECT 1 FROM sys.databases where name=@p1`
		rows, err := db.Query(statement, databaseName)
		if err != nil {
			return fmt.Errorf("error querying existence of database %q: %w", databaseName, err)
		}
		defer rows.Close()

		exists := rows.Next()

		if !exists {
			return fmt.Errorf("database %s was not created", databaseName)
		}

		return nil
	}
}

func getTestProviderFactories(provider *schema.Provider) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		providerName: func() (*schema.Provider, error) {
			if provider == nil {
				return provider, errors.New("provider cannot be nil")
			}

			return provider, nil
		},
	}
}

func initTestProvider() *schema.Provider {
	testAccProvider := &schema.Provider{
		Schema: csbsqlserver.GetProviderSchema(),
		ResourcesMap: map[string]*schema.Resource{
			csbsqlserver.ResourceNameKey: csbsqlserver.BindingResource(),
		},
		ConfigureContextFunc: csbsqlserver.ProviderContextFunc,
	}
	err := testAccProvider.InternalValidate()
	Expect(err).NotTo(HaveOccurred())

	return testAccProvider
}

func testGetConfiguration(port int, adminPassword, bindingUser1, bindingPassword1, bindingUser2, bindingPassword2, databaseName string) string {
	return fmt.Sprintf(`
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
				roles    = ["db_ddladmin", "db_datareader", "db_datawriter", "db_accessadmin"]
			}

			resource "csbsqlserver_binding" "binding2" {
				username   = "%s"
				password   = "%s"
				roles      = ["db_ddladmin", "db_datareader", "db_datawriter", "db_accessadmin"]
                depends_on = [csbsqlserver_binding.binding1]
			}`,
		testhelpers.Server,
		port,
		databaseName,
		testhelpers.AdminUser,
		adminPassword,
		bindingUser1,
		bindingPassword1,
		bindingUser2,
		bindingPassword2,
	)
}

func testGetConfigurationOnlyBindingOne(port int, adminPassword, bindingUser1, bindingPassword1, databaseName string) string {
	return fmt.Sprintf(`
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
				roles    = ["db_ddladmin", "db_datareader", "db_datawriter", "db_accessadmin"]
			}`,
		testhelpers.Server,
		port,
		databaseName,
		testhelpers.AdminUser,
		adminPassword,
		bindingUser1,
		bindingPassword1,
	)
}
