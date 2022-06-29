package testhelpers

import (
	"database/sql"
	"fmt"

	"github.com/onsi/gomega"
)

func UserExists(db *sql.DB, username string) bool {
	rows, err := db.Query(`SELECT NAME FROM sys.database_principals WHERE NAME = @p1 AND TYPE = 'S'`, username)
	gomega.Expect(err).WithOffset(1).NotTo(gomega.HaveOccurred())
	defer rows.Close()
	return rows.Next()
}

// execf does an Exec with Printf style parameters
func execf(db *sql.DB, format string, a ...any) {
	_, err := db.Exec(fmt.Sprintf(format, a...))
	gomega.Expect(err).WithOffset(1).NotTo(gomega.HaveOccurred())
}
