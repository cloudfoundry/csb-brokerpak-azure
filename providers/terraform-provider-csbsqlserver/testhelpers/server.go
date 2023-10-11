package testhelpers

import (
	"database/sql"
	"fmt"
	"os/exec"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	AdminUser       = "sa"
	Server          = "localhost"
	TestDatabase    = "testdb"
	defaultDatabase = "tempdb"
)

type ShutdownServerFn func(time.Duration)

func StartServer(password string, port int, opts ...ServerOption) ShutdownServerFn {
	// The SA_PASSWORD environment variable is deprecated. Use MSSQL_SA_PASSWORD instead.

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-e", "ACCEPT_EULA=Y", // EULA allows usage for testing
		"-e", fmt.Sprintf("MSSQL_SA_PASSWORD=%s", password),
		"-p", fmt.Sprintf("%d:1433", port),
		"mcr.microsoft.com/mssql/server:2022-latest",
	)
	shutdownServerFn := runServerOrFail(password, port, cmd)

	db := ConnectDefaultDB(AdminUser, password, port)

	for _, opt := range opts {
		opt(db)
	}

	if len(opts) == 0 {
		WithSPConfigure()(db)
		WithDatabase()(db)
	}

	return shutdownServerFn
}

func runServerOrFail(password string, port int, cmd *exec.Cmd) ShutdownServerFn {
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	for ready, started := false, time.Now(); !ready; {
		db := Connect(AdminUser, password, defaultDatabase, port)

		switch {
		case db.Ping() == nil: // successful ping
			ready = true
		case session.ExitCode() != -1: // docker image no longer running
			ginkgo.Fail("server running in docker has exited")
		case time.Since(started) > 5*time.Minute:
			ginkgo.Fail("timed out waiting for the server to start")
		default:
			time.Sleep(time.Second)
		}
	}
	return func(duration time.Duration) {
		session.Terminate().Wait(duration)
	}
}

type ServerOption func(db *sql.DB)

func WithSPConfigure() ServerOption {
	return func(db *sql.DB) {
		execf(db, `EXEC sp_configure 'contained database authentication', 1 RECONFIGURE`)
	}
}

func WithDatabase() ServerOption {
	return func(db *sql.DB) {
		execf(db, `CREATE DATABASE %s CONTAINMENT = PARTIAL`, TestDatabase)
	}
}

func WithNoop() ServerOption { return func(db *sql.DB) {} }

func Connect(username, password, database string, port int) *sql.DB {
	db, err := sql.Open("sqlserver", NewEncoder(Server, username, password, database, "disable", port).Encode())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return db
}

func ConnectDefaultDB(username, password string, port int) *sql.DB {
	db, err := sql.Open("sqlserver", NewEncoder(Server, username, password, "", "disable", port).EncodeWithoutDB())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return db
}
