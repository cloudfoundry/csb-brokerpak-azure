// Package connector manages the creating and deletion of service bindings to MS SQL Server
package connector

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
)

const (
	roleName = "binding_user_group"
)

func New(server string, port int, username, password, database, encrypt string) *Connector {
	return &Connector{
		database: database,
		encoder: NewEncoder(
			server,
			username,
			password,
			database,
			encrypt,
			port,
		)}
}

type Connector struct {
	database string
	encoder  *Encoder
}

// CreateBinding creates the binding user, adds roles and grants permission to execute store procedures
// It is idempotent.
func (c *Connector) CreateBinding(ctx context.Context, username, password string, roles []string) error {
	exists, err := c.CheckDatabaseExists(ctx, c.database)
	if err != nil {
		return err
	}

	if !exists {
		// First binding => future provision phase
		// Transact-SQL Statements not allowed in a Transaction:
		// https://learn.microsoft.com/en-us/previous-versions/sql/sql-server-2008-r2/ms191544(v=sql.105)
		if err := c.createDatabase(ctx, c.database); err != nil {
			return err
		}

		if err := c.setAutoClose(ctx, c.database); err != nil {
			return err
		}
	}

	return c.withTransaction(func(tx *sql.Tx) error {

		if err := createUser(ctx, tx, username, password); err != nil {
			return err
		}

		if err := addRoles(ctx, tx, username, roles); err != nil {
			return err
		}

		if err := grantExec(ctx, tx, username); err != nil {
			return err
		}

		return nil
	})
}

// DeleteBinding drops the binding user. It is idempotent.
func (c *Connector) DeleteBinding(ctx context.Context, username string) error {

	if err := c.ManageObjectReassignment(ctx, username); err != nil {
		return err
	}

	return c.withTransaction(func(tx *sql.Tx) error {
		if err := dropUser(ctx, tx, username); err != nil {
			return err
		}
		if err := dropLogin(ctx, tx, username); err != nil {
			return err
		}

		return nil
	})
}

// ReadBinding checks whether the binding exists
func (c *Connector) ReadBinding(ctx context.Context, username string) (result bool, err error) {
	return result, c.withConnection(func(db *sql.DB) (err error) {
		result, err = checkUser(ctx, db, username)
		return
	})
}

func (c *Connector) CheckDatabaseExists(ctx context.Context, dbName string) (result bool, err error) {
	return result, c.withDefaultDBConnection(func(db *sql.DB) (err error) {
		statement := `SELECT 1 FROM sys.databases where name=@p1`
		rows, err := db.QueryContext(ctx, statement, dbName)
		if err != nil {
			return fmt.Errorf("error querying existence of database %q: %w", dbName, err)
		}
		defer rows.Close()

		result = rows.Next()
		return nil
	})
}

func (c *Connector) ManageObjectReassignment(ctx context.Context, username string) error {
	exists, err := c.checkObjectReassignmentRoleExist(ctx)
	if err != nil {
		return err
	}

	if !exists {
		if err := c.createObjectReassignmentRole(ctx); err != nil {
			return err
		}
	}

	if err := c.addUserToObjectReassignmentRole(ctx, username); err != nil {
		return err
	}

	if err := c.transferOwnershipToObjectReassignmentRole(ctx, username); err != nil {
		return err
	}

	return nil
}

func (c *Connector) checkObjectReassignmentRoleExist(ctx context.Context) (result bool, err error) {
	return result, c.withConnection(func(db *sql.DB) (err error) {
		statement := `SELECT 1 FROM sys.database_principals WHERE type_desc = 'DATABASE_ROLE' AND name = @p1`
		rows, err := db.QueryContext(ctx, statement, roleName)
		if err != nil {
			return fmt.Errorf("error querying existence of role %q: %w", roleName, err)
		}
		defer rows.Close()

		result = rows.Next()
		return nil
	})
}

func (c *Connector) createObjectReassignmentRole(ctx context.Context) error {
	return c.withConnection(func(db *sql.DB) (err error) {
		// It needs to be granted to the **user** database and not the **master**.
		// We change the database context through the connection.
		// We must use the specific database connection!! (withConnection instead of withDefaultDBConnection)
		// otherwise when trying to add the user to the role, we will not have permissions.
		// Remember that the user authentication information is stored in each database.
		statement := fmt.Sprintf("CREATE ROLE %s AUTHORIZATION dbo", roleName)
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("error creating role %q: %w", roleName, err)
		}
		return nil
	})
}

func (c *Connector) addUserToObjectReassignmentRole(ctx context.Context, username string) error {
	return c.withConnection(func(db *sql.DB) (err error) {
		statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'ALTER ROLE ' + QuoteName(@roleName) + ' ADD MEMBER ' + QuoteName(@username) + ' '
EXEC (@sql)
`
		if _, err := db.ExecContext(ctx, statement, sql.Named("roleName", roleName), sql.Named("username", username)); err != nil {
			return fmt.Errorf("error adding user %q to role %q: %w", username, roleName, err)
		}
		return nil
	})
}

func (c *Connector) transferOwnershipToObjectReassignmentRole(ctx context.Context, username string) error {
	statement := `
DECLARE @cur CURSOR, @sql nvarchar(max), @schemaName nvarchar(max);
SET @cur = CURSOR STATIC FOR
    SELECT schema_name FROM information_schema.schemata WHERE schema_owner=@username

OPEN @cur

WHILE 1 = 1
BEGIN
     FETCH @cur INTO @schemaName
     IF @@fetch_status <> 0
        BREAK
     SET @sql = 'ALTER AUTHORIZATION ON SCHEMA::' + QuoteName(@schemaName) + ' TO ' + QuoteName(@roleName) + ' '
	 EXEC (@sql)
END
`

	return c.withTransaction(func(tx *sql.Tx) (err error) {
		if _, err := tx.ExecContext(ctx, statement, sql.Named("username", username), sql.Named("roleName", roleName)); err != nil {
			return fmt.Errorf("error transfering ownership of schemas owned by %s to the %s role: %w", username, roleName, err)
		}
		return nil
	})
}

func (c *Connector) createDatabase(ctx context.Context, dbName string) error {
	statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'CREATE DATABASE ' + QuoteName(@databaseName) + ' CONTAINMENT=PARTIAL' 
EXEC (@sql)
`
	return c.withDefaultDBConnection(func(db *sql.DB) (err error) {
		if _, err := db.ExecContext(ctx, statement, sql.Named("databaseName", dbName)); err != nil {
			return fmt.Errorf("error creating database %q: %w", dbName, err)
		}
		return nil
	})
}

func (c *Connector) setAutoClose(ctx context.Context, dbName string) error {
	statement := `
DECLARE @sql nvarchar(max)
SET @sql = 'ALTER DATABASE ' + QuoteName(@databaseName) + ' SET AUTO_CLOSE OFF' 
EXEC (@sql)
`
	return c.withDefaultDBConnection(func(db *sql.DB) (err error) {
		if _, err := db.ExecContext(ctx, statement, sql.Named("databaseName", dbName)); err != nil {
			return fmt.Errorf("error setting autoclose %q: %w", dbName, err)
		}
		return nil
	})
}

// withConnection returns a DB connection tied to the user's database connection
func (c *Connector) withConnection(callback func(*sql.DB) error) error {
	db, err := c.connect(c.encoder.Encode())
	if err != nil {
		return err
	}

	return callback(db)
}

// withDefaultDBConnection returns a DB connection tied to the default database of the server
func (c *Connector) withDefaultDBConnection(callback func(*sql.DB) error) error {
	db, err := c.connect(c.encoder.EncodeWithoutDB())
	if err != nil {
		return err
	}

	return callback(db)
}

func (c *Connector) withTransaction(callback func(*sql.Tx) error) error {
	return c.withConnection(func(db *sql.DB) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		if err := callback(tx); err != nil {
			return err
		}

		return tx.Commit()
	})
}

func (c *Connector) connect(url string) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", url)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

type databaseActor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func checkUser(ctx context.Context, db databaseActor, username string) (bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT NAME FROM sys.database_principals WHERE NAME = @p1 AND TYPE = 'S'`, username)
	if err != nil {
		return false, fmt.Errorf("error querying existence of user %q: %w", username, err)
	}
	defer rows.Close()
	return rows.Next(), nil
}

func dropUser(ctx context.Context, tx *sql.Tx, username string) error {
	statement := fmt.Sprintf(`DROP USER IF EXISTS [%s]`, username)
	if _, err := tx.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error deleting user %q: %w", username, err)
	}

	return nil
}

// dropLogin drops a legacy login. In the past the brokerpak created a login and
// then created a user from the login, but this was problematic in a failover because
// the login did not exist on the replica. See: https://www.pivotaltracker.com/story/show/179168006
func dropLogin(ctx context.Context, tx *sql.Tx, username string) error {
	// Unfortunately there's no "DROP LOGIN IF EXISTS" and checking for the existence of the login (via sys.sql_logins)
	// proved unreliable as it worked on the database fake, but not on an Azure database. So we just attempt the operation
	// and ignore any error. The issue with checking for the "right" error is that any check would be fragile if there were
	// future changes to the message, SQLState, etc... and the benefit was not considered to be worth the risk
	_, _ = tx.ExecContext(ctx, fmt.Sprintf(`DROP LOGIN [%s]`, username))

	return nil
}

func createUser(ctx context.Context, tx *sql.Tx, username, password string) error {
	ok, err := checkUser(ctx, tx, username)
	switch {
	case err != nil:
		return err
	case ok:
		return nil // already exists. In the future we could update the password with "ALTER USER".
	}

	statement := fmt.Sprintf(`CREATE USER [%s] with PASSWORD='%s'`, username, password)
	if _, err := tx.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error creating user %q: %w", username, err)
	}

	return nil
}

func addRoles(ctx context.Context, tx *sql.Tx, username string, roles []string) error {
	for _, role := range roles {
		statement := fmt.Sprintf(`ALTER ROLE %s ADD MEMBER [%s]`, role, username)
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("error adding user %q to role %q: %w", username, role, err)
		}
	}

	return nil
}

// grantExec grants EXECUTE permission to all the user stored procedures.
// Right out of the box, SQL Server makes it pretty easy to grant SELECT, INSERT, UPDATE, and DELETE to all user tables.
// That’s accomplished by using the built-in db_datareader (SELECT) and
// db_datawriter (INSERT, UPDATE, and DELETE) database roles in every user database.
// Any user you add to those database roles will be granted those permissions.
// But what if you want to grant EXECUTE permission to all the user stored procedures.
// Where’s the built-in database role for that? Nowhere to be found.
func grantExec(ctx context.Context, tx *sql.Tx, username string) error {
	statement := fmt.Sprintf(`GRANT EXEC TO [%s]`, username)
	if _, err := tx.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error granting exec to user %q: %w", username, err)
	}

	return nil
}
