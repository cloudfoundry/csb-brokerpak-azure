package csbsqlserver_test

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	. "github.com/onsi/ginkgo/v2"

	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbsqlserver/testhelpers"
)

var _ = Describe("csbsqlserver_binding resource", func() {

	Context("database does not exists", func() {
		When("binding is deleted", func() {
			It("should allow access to previous objects", func() {
				var (
					adminPassword = testhelpers.RandomPassword()
					port          = testhelpers.FreePort()
				)

				shutdownServerFn := testhelpers.StartServer(adminPassword, port, testhelpers.WithSPConfigure())
				DeferCleanup(func() { shutdownServerFn(time.Minute) })

				cnf := createTestCaseCnf(adminPassword, port)
				tableNameUserOne := testhelpers.RandomTableName()
				tableNameUserTwo := testhelpers.RandomTableName()

				schemaNameUserOne := testhelpers.RandomSchemaName("user_one")
				schemaNameUserTwo := testhelpers.RandomSchemaName("user_two")

				resource.Test(GinkgoT(),
					getTestCase(cnf,
						getMandatoryStep(cnf,
							// user ONE reads and writes content in a table created by himself
							testCheckUserCanCreateTable(cnf, tableNameUserOne, cnf.ResourceBindingOneName),
							testCheckUserCanWriteContent(cnf, tableNameUserOne, cnf.ResourceBindingOneName, "fake_content"),
							testCheckUserCanReadContent(cnf, tableNameUserOne, cnf.ResourceBindingOneName, "fake_content"),
							// user TWO reads and writes content in a table created by himself
							testCheckUserCanCreateTable(cnf, tableNameUserTwo, cnf.ResourceBindingTwoName),
							testCheckUserCanWriteContent(cnf, tableNameUserTwo, cnf.ResourceBindingTwoName, "fake_content"),
							testCheckUserCanReadContent(cnf, tableNameUserTwo, cnf.ResourceBindingTwoName, "fake_content"),
							// user TWO reads content in the table created by user ONE
							testCheckUserCanReadContent(cnf, tableNameUserOne, cnf.ResourceBindingTwoName, "fake_content"),
							// user ONE reads content in the table created by user TWO
							testCheckUserCanReadContent(cnf, tableNameUserTwo, cnf.ResourceBindingOneName, "fake_content"),
							// user ONE creates schema
							testCheckUserCanCreateSchema(cnf, schemaNameUserOne, cnf.ResourceBindingOneName),
							testCheckEntityIsOwnerOfTheSchema(cnf, cnf.ResourceBindingOneName, cnf.BindingUserOne, schemaNameUserOne),
							testCheckUserIsNotOwnerOfTheSchema(cnf, cnf.ResourceBindingTwoName, schemaNameUserTwo),
							// user ONE creates table in schema, write and read content
							testCheckUserCanCreateTableInSchema(cnf, schemaNameUserOne, tableNameUserOne, cnf.ResourceBindingOneName),
							testCheckUserCanWriteContentInSchema(cnf, schemaNameUserOne, tableNameUserOne, cnf.ResourceBindingOneName, "fake_content"),
							testCheckUserCanReadContentInSchema(cnf, schemaNameUserOne, tableNameUserOne, cnf.ResourceBindingOneName, "fake_content"),
							// user TWO creates schema
							testCheckUserCanCreateSchema(cnf, schemaNameUserTwo, cnf.ResourceBindingTwoName),
							testCheckEntityIsOwnerOfTheSchema(cnf, cnf.ResourceBindingTwoName, cnf.BindingUserTwo, schemaNameUserTwo),
							testCheckUserIsNotOwnerOfTheSchema(cnf, cnf.ResourceBindingOneName, schemaNameUserTwo),
							// user TWO creates table in schema, write and read content
							testCheckUserCanCreateTableInSchema(cnf, schemaNameUserTwo, tableNameUserTwo, cnf.ResourceBindingTwoName),
							testCheckUserCanWriteContentInSchema(cnf, schemaNameUserTwo, tableNameUserTwo, cnf.ResourceBindingTwoName, "fake_content"),
							testCheckUserCanReadContentInSchema(cnf, schemaNameUserTwo, tableNameUserTwo, cnf.ResourceBindingTwoName, "fake_content"),
							// user ONE reads schema names
							testCheckSchemaNames(cnf, cnf.ResourceBindingOneName, schemaNameUserOne, schemaNameUserTwo),
							// user TWO reads schema names
							testCheckSchemaNames(cnf, cnf.ResourceBindingTwoName, schemaNameUserOne, schemaNameUserTwo),
							// user ONE CANNOT create a database
							testCheckCreateDatabase(cnf, "fake_database_name", cnf.ResourceBindingOneName, `mssql: CREATE DATABASE permission denied in database 'master'`),
						),
						getStepOnlyBindingOne(cnf,
							// user ONE reads content in a table created by himself - user TWO was deleted
							testCheckUserCanReadContent(cnf, tableNameUserOne, cnf.ResourceBindingOneName, "fake_content"),
							// user ONE reads content in a table created by USER TWO - user TWO was deleted
							testCheckUserCanReadContent(cnf, tableNameUserTwo, cnf.ResourceBindingOneName, "fake_content"),
							// user ONE writes content in a table created by USER TWO - user TWO was deleted
							testCheckUserCanWriteContent(cnf, tableNameUserTwo, cnf.ResourceBindingOneName, "another_content"),
							// user ONE reads content created by himself in a table created by USER TWO - user TWO was deleted
							testCheckUserCanReadContent(cnf, tableNameUserTwo, cnf.ResourceBindingOneName, "another_content"),
							// user ONE reads content created by himself in a schema.table created by himself - user TWO was deleted
							testCheckUserCanReadContentInSchema(cnf, schemaNameUserOne, tableNameUserOne, cnf.ResourceBindingOneName, "fake_content"),
							// user ONE is not the owner of the schema created by user TWO - user TWO was deleted
							testCheckUserIsNotOwnerOfTheSchema(cnf, cnf.ResourceBindingOneName, schemaNameUserTwo),
							// role "binding_user_group" is the owner of the schema created by user TWO - user TWO was deleted
							testCheckEntityIsOwnerOfTheSchema(cnf, cnf.ResourceBindingOneName, "binding_user_group", schemaNameUserTwo),
							// user ONE reads content created by USER TWO in a schema.table created by USER TWO - user TWO was deleted
							testCheckUserCanReadContentInSchema(cnf, schemaNameUserTwo, tableNameUserTwo, cnf.ResourceBindingOneName, "fake_content"),
							// user ONE writes content in a schema.table created by USER TWO - user TWO was deleted
							testCheckUserCanWriteContentInSchema(cnf, schemaNameUserTwo, tableNameUserTwo, cnf.ResourceBindingOneName, "fake_content"),
							// user ONE reads schema names
							testCheckSchemaNames(cnf, cnf.ResourceBindingOneName, schemaNameUserOne, schemaNameUserTwo),
							// user ONE drops table in the schema created by himself
							testCheckDropTableInSchema(cnf, schemaNameUserOne, tableNameUserOne, cnf.ResourceBindingOneName),
							// user ONE drops the schema created by himself
							testCheckDroopSchema(cnf, schemaNameUserOne, cnf.ResourceBindingOneName),
							// user ONE reads the schema created by himself - schema does not exist
							func(state *terraform.State) error {
								err := testCheckSchemaNames(cnf, cnf.ResourceBindingOneName, schemaNameUserOne)(state)
								if err != nil && strings.Contains(err.Error(), "schema not found") {
									return nil
								}
								return err
							},
							// user ONE reads the schema created by USER TWO - user TWO was deleted
							testCheckSchemaNames(cnf, cnf.ResourceBindingOneName, schemaNameUserTwo),
							// user ONE creates table in schema created by USER TWO - user TWO was deleted
							testCheckUserCanCreateTableInSchema(cnf, schemaNameUserTwo, tableNameUserOne, cnf.ResourceBindingOneName),
							// user ONE drops table created by himself in the schema created by USER TWO
							testCheckDropTableInSchema(cnf, schemaNameUserTwo, tableNameUserOne, cnf.ResourceBindingOneName),
							// user ONE alter table in schema created by user TWO
							testCheckAddColumnToTableInSchema(cnf, schemaNameUserTwo, tableNameUserTwo, cnf.ResourceBindingOneName, "another_column"),
							// user ONE drops table created by USER TWO in the schema created by USER TWO
							testCheckDropTableInSchema(cnf, schemaNameUserTwo, tableNameUserTwo, cnf.ResourceBindingOneName),
							// user ONE drops the schema created by USER TWO
							testCheckDroopSchema(cnf, schemaNameUserTwo, cnf.ResourceBindingOneName),
						),
					),
				)
			})
		})
	})
})

type credentials struct {
	username, password string
}

func testCheckUserCanCreateTable(cnf testCaseCnf, randomTableName, resourceBindingName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		_, err = db.Exec(getCreateTableSQL(randomTableName))
		if err != nil {
			return fmt.Errorf("error creating table %w", err)
		}

		return nil
	}
}

func testCheckUserCanCreateSchema(cnf testCaseCnf, randomSchemaName, resourceBindingName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		_, err = db.Exec(getCreateSchemaSQL(randomSchemaName, cr.username))
		if err != nil {
			return fmt.Errorf("error creating schema %w", err)
		}

		return nil
	}
}

func testCheckUserCanCreateTableInSchema(cnf testCaseCnf, schemaName, randomTableName, resourceBindingName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		name := fmt.Sprintf("%s.%s", schemaName, randomTableName)
		_, err = db.Exec(getCreateTableSQL(name))
		if err != nil {
			return fmt.Errorf("error creating table %w", err)
		}

		return nil
	}
}

func testCheckUserCanWriteContent(cnf testCaseCnf, randomTableName, resourceBindingName, fakeContent string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		result, err := db.Exec(getInsertRowSQL(randomTableName), fakeContent)
		if err != nil {
			return fmt.Errorf("error inserting row %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if int(rowsAffected) != 1 {
			return fmt.Errorf("error creating row")
		}

		return nil
	}
}

func testCheckUserCanReadContent(cnf testCaseCnf, randomTableName, resourceBindingName, expectedFakeContent string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		var v string
		if err := db.QueryRow(getValueSQL(randomTableName), expectedFakeContent).Scan(&v); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("value not found %w", err)
			}
			return fmt.Errorf("error retrieving value %w", err)
		}

		return nil
	}
}

func testCheckUserCanWriteContentInSchema(cnf testCaseCnf, schemaName, randomTableName, resourceBindingName, fakeContent string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		name := fmt.Sprintf("%s.%s", schemaName, randomTableName)
		result, err := db.Exec(getInsertRowSQL(name), fakeContent)
		if err != nil {
			return fmt.Errorf("error inserting row %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if int(rowsAffected) != 1 {
			return fmt.Errorf("error creating row")
		}

		return nil
	}
}

func testCheckUserCanReadContentInSchema(cnf testCaseCnf, schemaName, randomTableName, resourceBindingName, expectedFakeContent string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		name := fmt.Sprintf("%s.%s", schemaName, randomTableName)
		var v string
		if err := db.QueryRow(getValueSQL(name), expectedFakeContent).Scan(&v); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("value not found %w", err)
			}
			return fmt.Errorf("error retrieving value %w", err)
		}

		return nil
	}
}

func testCheckSchemaNames(cnf testCaseCnf, resourceBindingName string, expectedSchemaNames ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		rows, err := db.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA")
		if err != nil {
			return fmt.Errorf("error retrieving schemas %w", err)
		}
		defer rows.Close()

		var schemas []string

		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				return fmt.Errorf("error scanning schema %w", err)
			}
			schemas = append(schemas, s)
		}

		if rows.Err() != nil {
			return rows.Err()
		}

		var multiErr error
	external:
		for _, expectedSchemaName := range expectedSchemaNames {

			for _, schema := range schemas {
				if schema == expectedSchemaName {
					continue external
				}
			}

			err := fmt.Errorf("schema not found: %s", expectedSchemaName)
			if multiErr != nil {
				multiErr = errors.Join(multiErr, err)
			} else {
				multiErr = err
			}
		}

		if multiErr != nil {
			return fmt.Errorf("error checking schemas - available schemas %s: %w", strings.Join(schemas, ","), multiErr)
		}

		return multiErr
	}
}

func testCheckUserIsNotOwnerOfTheSchema(cnf testCaseCnf, resourceBindingName string, expectedSchemaName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		found, err := findSchemaByOwner(db, cr.username, expectedSchemaName)
		if err != nil {
			return err
		}

		if found {
			return fmt.Errorf("the user %s is owner of the schema: %s", cr.username, expectedSchemaName)
		}

		return nil
	}
}

func testCheckEntityIsOwnerOfTheSchema(cnf testCaseCnf, resourceBindingName string, roleName, expectedSchemaName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		found, err := findSchemaByOwner(db, roleName, expectedSchemaName)
		if err != nil {
			return err
		}

		if !found {
			return fmt.Errorf("the role %s is not the owner of the schema: %s", roleName, expectedSchemaName)
		}

		return nil
	}
}

func testCheckDropTableInSchema(cnf testCaseCnf, schemaName, randomTableName, resourceBindingName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		if _, err := db.Exec(fmt.Sprintf("DROP TABLE %s.%s", schemaName, randomTableName)); err != nil {
			return fmt.Errorf("error dropping table %w", err)
		}

		return nil
	}
}

func testCheckDroopSchema(cnf testCaseCnf, schemaName, resourceBindingName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		if _, err := db.Exec(fmt.Sprintf("DROP SCHEMA %s", schemaName)); err != nil {
			return fmt.Errorf("error deleting schema %w", err)
		}

		return nil
	}
}

func testCheckCreateDatabase(cnf testCaseCnf, dbName, resourceBindingName, expectedError string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'CREATE DATABASE ' + QuoteName(@databaseName) + ' CONTAINMENT=PARTIAL' 
EXEC (@sql)
`
		if _, err := db.Exec(statement, sql.Named("databaseName", dbName)); err != nil {
			if !strings.Contains(err.Error(), expectedError) {
				return fmt.Errorf("unexpected error %w", err)
			}
		}

		return nil
	}
}

func testCheckAddColumnToTableInSchema(cnf testCaseCnf, schemaName, randomTableName, resourceBindingName, columnName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cr, err := getCredentialsFromState(state, resourceBindingName)
		if err != nil {
			return err
		}

		db := testhelpers.Connect(cr.username, cr.password, cnf.DatabaseName, cnf.Port)
		defer db.Close()

		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s.%s ADD %s VARCHAR(20) NULL", schemaName, randomTableName, columnName)); err != nil {
			return fmt.Errorf("error altering table %w", err)
		}

		return nil
	}
}

func getCredentialsFromState(state *terraform.State, resourceBindingName string) (credentials, error) {
	rs, err := retrieveResourceByNameFromState(state, resourceBindingName)
	if err != nil {
		return credentials{}, err
	}

	username, err := getPropertyValueFromResource(rs, "username")
	if err != nil {
		return credentials{}, err
	}

	password, err := getPropertyValueFromResource(rs, "password")
	if err != nil {
		return credentials{}, err
	}

	return credentials{username: username, password: password}, nil
}

func retrieveResourceByNameFromState(state *terraform.State, resourceName string) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}
	return rs, nil
}

func getPropertyValueFromResource(rs *terraform.ResourceState, attrName string) (string, error) {
	v, ok := rs.Primary.Attributes[attrName]
	if !ok {
		return "", fmt.Errorf("attribute %q not found", attrName)
	}
	return v, nil
}

func findSchemaByOwner(db *sql.DB, entityOwner string, expectedSchemaName string) (bool, error) {
	rows, err := db.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_OWNER = @p1", entityOwner)
	if err != nil {
		return false, fmt.Errorf("error retrieving schemas %w", err)
	}
	defer rows.Close()

	var schemas []string

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return false, fmt.Errorf("error scanning schema %w", err)
		}
		schemas = append(schemas, s)
	}

	if rows.Err() != nil {
		return false, rows.Err()
	}

	found := false
	for _, schema := range schemas {
		if schema == expectedSchemaName {
			found = true
			break
		}
	}
	return found, nil
}

func getCreateTableSQL(tableName string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
    id INT PRIMARY KEY IDENTITY (1, 1),
    value VARCHAR (50) NOT NULL
)`, tableName)
}

func getInsertRowSQL(tableName string) string {
	return fmt.Sprintf(`
INSERT INTO %s(value) values(@p1)`, tableName)
}

func getValueSQL(tableName string) string {
	return fmt.Sprintf(`
SELECT value FROM %s WHERE value = @p1`, tableName)
}

func getCreateSchemaSQL(schemaName, user string) string {
	return fmt.Sprintf(`CREATE SCHEMA %s AUTHORIZATION [%s]`, schemaName, user)
}
