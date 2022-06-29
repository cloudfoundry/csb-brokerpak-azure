// Package connector manages the creating and deletion of service bindings to MS SQL Server
package connector

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
)

func New(server string, port int, username, password, database, encrypt string) *Connector {
	return &Connector{
		server:   server,
		username: username,
		password: password,
		database: database,
		encrypt:  encrypt,
		port:     port,
	}
}

type Connector struct {
	server, username, password, database, encrypt string
	port                                          int
}

// CreateBinding creates the binding user, adds roles and grants permission to execute store procedures
// It is idempotent.
func (c *Connector) CreateBinding(ctx context.Context, username, password string, roles []string) error {
	return c.withConnection(func(db *sql.DB) error {
		if err := createUser(ctx, db, username, password); err != nil {
			return err
		}

		if err := addRoles(ctx, db, username, roles); err != nil {
			return err
		}

		if err := grantExec(ctx, db, username); err != nil {
			return err
		}

		return nil
	})
}

// DeleteBinding drops the binding user. It is idempotent.
func (c *Connector) DeleteBinding(ctx context.Context, username string) error {
	return c.withConnection(func(db *sql.DB) error {
		if err := dropUser(ctx, db, username); err != nil {
			return err
		}
		if err := dropLogin(ctx, db, username); err != nil {
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

func (c *Connector) withConnection(callback func(db *sql.DB) error) error {
	db, err := sql.Open("sqlserver", c.connStr())
	if err != nil {
		return fmt.Errorf("error connecting to database %q on %q port %d with user %q: %w", c.database, c.server, c.port, c.username, err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("error pinging database %q on %q port %d with user %q: %w", c.database, c.server, c.port, c.username, err)
	}

	return callback(db)
}

func (c *Connector) connStr() string {
	return strings.Join(
		[]string{
			fmt.Sprintf("server=%s", c.server),
			fmt.Sprintf("user id=%s", c.username),
			fmt.Sprintf("password=%s", c.password),
			fmt.Sprintf("port=%d", c.port),
			fmt.Sprintf("database=%s", c.database),
			fmt.Sprintf("encrypt=%s", c.encrypt),
		},
		";",
	)
}

func checkUser(ctx context.Context, db *sql.DB, username string) (bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT NAME FROM sys.database_principals WHERE NAME = @p1 AND TYPE = 'S'`, username)
	if err != nil {
		return false, fmt.Errorf("error querying existence of user %q: %w", username, err)
	}
	defer rows.Close()
	return rows.Next(), nil
}

func dropUser(ctx context.Context, db *sql.DB, username string) error {
	statement := fmt.Sprintf(`DROP USER IF EXISTS [%s]`, username)
	if _, err := db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error deleting user %q: %w", username, err)
	}

	return nil
}

// dropLogin drops a legacy login. In the past the brokerpak created a login and
// then created a user from the login, but this was problematic in a failover because
// the login did not exist on the replica. See: https://www.pivotaltracker.com/story/show/179168006
func dropLogin(ctx context.Context, db *sql.DB, username string) error {
	// Unfortunately there's no "DROP LOGIN IF EXISTS" and checking for the existence of the login (via sys.sql_logins)
	// proved unreliable as it worked on the database fake, but not on an Azure database. So we just attempt the operation
	// and ignore any error. The issue with checking for the "right" error is that any check would be fragile if there were
	// future changes to the message, SQLState, etc... and the benefit was not considered to be worth the risk
	_, _ = db.ExecContext(ctx, fmt.Sprintf(`DROP LOGIN [%s]`, username))

	return nil
}

func createUser(ctx context.Context, db *sql.DB, username, password string) error {
	ok, err := checkUser(ctx, db, username)
	switch {
	case err != nil:
		return err
	case ok:
		return nil // already exists. In the future we could update the password with "ALTER USER".
	}

	statement := fmt.Sprintf(`CREATE USER [%s] with PASSWORD='%s'`, username, password)
	if _, err := db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error creating user %q: %w", username, err)
	}

	return nil
}

func addRoles(ctx context.Context, db *sql.DB, username string, roles []string) error {
	for _, role := range roles {
		statement := fmt.Sprintf(`ALTER ROLE %s ADD MEMBER [%s]`, role, username)
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("error adding user %q to role %q: %w", username, role, err)
		}
	}

	return nil
}

func grantExec(ctx context.Context, db *sql.DB, username string) error {
	statement := fmt.Sprintf(`GRANT EXEC TO [%s]`, username)
	if _, err := db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("error granting exec to user %q: %w", username, err)
	}

	return nil
}
