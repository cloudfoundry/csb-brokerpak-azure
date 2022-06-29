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

	gomega.Eventually(func(g gomega.Gomega) {
		db, err := sql.Open("sqlserver", ConnectionString(AdminUser, password, defaultDatabase, port))
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(db.Ping()).NotTo(gomega.HaveOccurred())
	}).WithTimeout(time.Minute).Should(gomega.Succeed())

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
