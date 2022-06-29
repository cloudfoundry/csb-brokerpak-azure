package testhelpers

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
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

func StartServer(password string, port int) *gexec.Session {
	cmd := exec.Command(
		"docker", "run",
		"-e", "ACCEPT_EULA=y", // EULA allows usage for testing
		"-e", fmt.Sprintf("SA_PASSWORD=%s", password),
		"-p", fmt.Sprintf("%d:1433", port),
		"mcr.microsoft.com/mssql/server:2019-latest",
	)
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	for ready, started := false, time.Now(); !ready; {
		db, err := sql.Open("sqlserver", ConnectionString(AdminUser, password, defaultDatabase, port))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		switch {
		case db.Ping() == nil: // successful ping
			ready = true
		case session.ExitCode() != -1: // docker image no longer running
			ginkgo.Fail("server running in docker has exited")
		case time.Since(started) > 10*time.Minute:
			ginkgo.Fail("timed out waiting for the server to start")
		default:
			time.Sleep(time.Second)
		}
	}

	db := Connect(AdminUser, password, defaultDatabase, port)
	execf(db, `EXEC sp_configure 'contained database authentication', 1 RECONFIGURE`)
	execf(db, `CREATE DATABASE %s CONTAINMENT = PARTIAL`, TestDatabase)

	return session
}

func ConnectionString(username, password, database string, port int) string {
	return strings.Join(
		[]string{
			fmt.Sprintf("server=%s", Server),
			fmt.Sprintf("user id=%s", username),
			fmt.Sprintf("password=%s", password),
			fmt.Sprintf("port=%d", port),
			fmt.Sprintf("database=%s", database),
			"encrypt=disable",
		},
		";",
	)
}

func Connect(username, password, database string, port int) *sql.DB {
	db, err := sql.Open("sqlserver", ConnectionString(username, password, database, port))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return db
}
