package connector_test

import (
	"context"
	"fmt"

	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/testhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

var _ = Describe("Connector", func() {
	Describe("CreateBinding()", func() {
		It("creates a binding", func() {
			bindingUsername := uuid.New()
			bindingPassword := testhelpers.RandomPassword()

			By("creating the binding")
			err := conn.CreateBinding(context.TODO(), bindingUsername, bindingPassword, []string{"db_accessadmin", "db_datareader"})
			Expect(err).NotTo(HaveOccurred())

			By("checking that the binding user exists")
			Expect(testhelpers.UserExists(db, bindingUsername)).To(BeTrue())

			By("checking that the binding user has the specified roles")
			Expect(userRoles(db, bindingUsername)).To(ContainElements("db_accessadmin", "db_datareader"))

			By("checking that the binding user can execute stored procedures")
			Expect(userPermissions(db, bindingUsername)).To(ContainElement("EXECUTE"))

			By("checking that it can connect and create data")
			udb := testhelpers.Connect(bindingUsername, bindingPassword, testhelpers.TestDatabase, port)
			_, err = udb.Exec(`CREATE SCHEMA test AUTHORIZATION dbo`)
			Expect(err).NotTo(HaveOccurred())
		})

		It("is idempotent", func() {
			bindingUsername := uuid.New()
			bindingPassword := testhelpers.RandomPassword()

			By("creating the binding")
			err := conn.CreateBinding(context.TODO(), bindingUsername, bindingPassword, []string{"db_accessadmin", "db_datareader"})
			Expect(err).NotTo(HaveOccurred())

			By("checking that it doesn't fail when created again")
			err = conn.CreateBinding(context.TODO(), bindingUsername, bindingPassword, []string{"db_accessadmin", "db_datareader"})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("DeleteBinding()", func() {
		It("removes a binding", func() {
			bindingUsername := uuid.New()
			bindingPassword := testhelpers.RandomPassword()

			By("creating the binding")
			err := conn.CreateBinding(context.TODO(), bindingUsername, bindingPassword, []string{"db_accessadmin", "db_datareader"})
			Expect(err).NotTo(HaveOccurred())

			By("checking that the binding user exists")
			Expect(testhelpers.UserExists(db, bindingUsername)).To(BeTrue())

			By("deleting the binding")
			err = conn.DeleteBinding(context.TODO(), bindingUsername)
			Expect(err).NotTo(HaveOccurred())

			By("checking that the binding user does not exist")
			Expect(testhelpers.UserExists(db, bindingUsername)).To(BeFalse(), "binding user still exists")
		})

		It("is idempotent", func() {
			bindingUsername := uuid.New()

			By("deleting a binding that does not exist")
			err := conn.DeleteBinding(context.TODO(), bindingUsername)
			Expect(err).NotTo(HaveOccurred())
		})

		It("removes legacy logins", func() {
			bindingUsername := uuid.New()
			bindingPassword := testhelpers.RandomPassword()

			By("creating a legacy login")
			_, err := db.Exec(fmt.Sprintf(`CREATE LOGIN [%s] with PASSWORD='%s'`, bindingUsername, bindingPassword))
			Expect(err).NotTo(HaveOccurred())
			db.Exec(fmt.Sprintf(`CREATE USER [%s] from LOGIN %s`, bindingUsername, bindingUsername))
			Expect(err).NotTo(HaveOccurred())

			By("deleting the binding")
			err = conn.DeleteBinding(context.TODO(), bindingUsername)
			Expect(err).NotTo(HaveOccurred())

			By("checking the login does not exist")
			rows, err := db.Query(`SELECT NAME FROM sys.sql_logins WHERE NAME = @p1`, bindingUsername)
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()
			Expect(rows.Next()).To(BeFalse(), "login still exists")
		})

		It("reassigns ownership on deletion", func() {
			bindingUsername := uuid.New()
			bindingPassword := testhelpers.RandomPassword()

			By("creating the binding")
			err := conn.CreateBinding(context.TODO(), bindingUsername, bindingPassword, []string{"db_owner"})
			Expect(err).NotTo(HaveOccurred())

			By("checking that the binding user exists")
			Expect(testhelpers.UserExists(db, bindingUsername)).To(BeTrue())

			By("connecting and creating data")
			value := uuid.New()
			udb := testhelpers.Connect(bindingUsername, bindingPassword, testhelpers.TestDatabase, port)
			_, err = udb.Exec(`CREATE SCHEMA reassignment`)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec(`CREATE TABLE reassignment.test (keyname VARCHAR(255) NOT NULL, valuename VARCHAR(max) NOT NULL)`)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec(`INSERT INTO reassignment.test (keyname, valuename) VALUES ('saved', @p1)`, value)
			Expect(err).NotTo(HaveOccurred())

			By("deleting the binding")
			err = conn.DeleteBinding(context.TODO(), bindingUsername)
			Expect(err).NotTo(HaveOccurred())

			By("checking that the binding user does not exist")
			Expect(testhelpers.UserExists(db, bindingUsername)).To(BeFalse(), "binding user still exists")
		})
	})

	Describe("ReadBinding()", func() {
		It("can detect whether a user exists", func() {
			bindingUsername := uuid.New()
			bindingPassword := testhelpers.RandomPassword()

			By("creating the binding")
			err := conn.CreateBinding(context.TODO(), bindingUsername, bindingPassword, []string{"db_accessadmin", "db_datareader"})
			Expect(err).NotTo(HaveOccurred())

			By("finding the user that exists")
			found, err := conn.ReadBinding(context.TODO(), bindingUsername)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())

			By("failing to find a user that doesn't exist")
			found, err = conn.ReadBinding(context.TODO(), uuid.New())
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})
	})

	Describe("persisting data", func() {
		It("persists data between bindings", func() {
			bindingUsername1 := uuid.New()
			bindingPassword1 := testhelpers.RandomPassword()

			By("creating a first binding")
			err := conn.CreateBinding(context.TODO(), bindingUsername1, bindingPassword1, []string{"db_accessadmin", "db_datareader"})
			Expect(err).NotTo(HaveOccurred())

			By("connecting and creating data")
			value := uuid.New()
			udb := testhelpers.Connect(bindingUsername1, bindingPassword1, testhelpers.TestDatabase, port)
			_, err = udb.Exec(`CREATE SCHEMA persist AUTHORIZATION dbo`)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec(`CREATE TABLE persist.test (keyname VARCHAR(255) NOT NULL, valuename VARCHAR(max) NOT NULL)`)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.Exec(`INSERT INTO persist.test (keyname, valuename) VALUES ('saved', @p1)`, value)
			Expect(err).NotTo(HaveOccurred())

			By("deleting the binding")
			err = conn.DeleteBinding(context.TODO(), bindingUsername1)
			Expect(err).NotTo(HaveOccurred())

			By("creating another binding")
			bindingUsername2 := uuid.New()
			bindingPassword2 := testhelpers.RandomPassword()
			err = conn.CreateBinding(context.TODO(), bindingUsername2, bindingPassword2, []string{"db_accessadmin", "db_datareader"})
			Expect(err).NotTo(HaveOccurred())

			By("checking that the data is still there")
			udb = testhelpers.Connect(bindingUsername2, bindingPassword2, testhelpers.TestDatabase, port)
			rows, err := udb.Query(`SELECT valuename FROM persist.test WHERE keyname='saved'`)
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()
			Expect(rows.Next()).To(BeTrue(), "no rows returned")

			By("checking that the data matches")
			var result string
			Expect(rows.Scan(&result)).NotTo(HaveOccurred())
			Expect(result).To(Equal(value), "recovered data does not match saved data")
		})
	})
})
