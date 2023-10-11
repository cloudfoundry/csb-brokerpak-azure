package connector_test

import (
	"database/sql"
	"testing"
	"time"

	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/connector"
	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/testhelpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConnector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Connector Suite")
}

var (
	port          int
	adminPassword string
	db            *sql.DB
	conn          *connector.Connector
)

var _ = BeforeSuite(func() {
	adminPassword = testhelpers.RandomPassword()
	port = testhelpers.FreePort()
	session := testhelpers.StartServer(adminPassword, port)
	DeferCleanup(func() {
		session.Terminate().Wait(time.Minute)
	})
	db = testhelpers.Connect(testhelpers.AdminUser, adminPassword, testhelpers.TestDatabase, port)
	conn = connector.New(testhelpers.Server, port, testhelpers.AdminUser, adminPassword, testhelpers.TestDatabase, "disable")
})

func userRoles(db *sql.DB, username string) (result []string) {
	query := `
        SELECT DPR.name AS DatabaseRoleName
        FROM sys.database_role_members AS DRM  
        RIGHT OUTER JOIN sys.database_principals AS DPR  
        ON DRM.role_principal_id = DPR.principal_id  
        LEFT OUTER JOIN sys.database_principals AS DPO  
        ON DRM.member_principal_id = DPO.principal_id  
        WHERE DPR.type = 'R' AND DPO.type = 'S' AND DPO.NAME = @p1`
	rows, err := db.Query(query, username)
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	defer rows.Close()
	for rows.Next() {
		var s string
		Expect(rows.Scan(&s)).NotTo(HaveOccurred())
		result = append(result, s)
	}
	return result
}

func userPermissions(db *sql.DB, username string) (result []string) {
	query := `
        SELECT PERMS.permission_name
        FROM sys.database_permissions PERMS
        INNER JOIN sys.database_principals DP ON PERMS.grantee_principal_id = DP.principal_id 
        WHERE DP.name = @p1`
	rows, err := db.Query(query, username)
	Expect(err).WithOffset(1).NotTo(HaveOccurred())
	defer rows.Close()
	for rows.Next() {
		var s string
		Expect(rows.Scan(&s)).NotTo(HaveOccurred())
		result = append(result, s)
	}
	return result
}
