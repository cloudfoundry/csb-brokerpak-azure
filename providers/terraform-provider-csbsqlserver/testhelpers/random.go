// Package testhelpers contains some helpers shared between different test packages
package testhelpers

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/onsi/gomega"
)

func FreePort() int {
	listener, err := net.Listen("tcp", "localhost:0")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func RandomPassword() string {
	return randomWithPrefix("AaZz09~.")
}

func RandomDatabaseName() string {
	return randomWithPrefix("database")
}

func RandomTableName() string {
	return strings.ReplaceAll(randomWithPrefix("table"), "-", "_")
}

func RandomSchemaName(prefixes ...string) string {
	p := strings.Join(prefixes, "_")
	p = fmt.Sprintf("schema_%s", p)
	return strings.ReplaceAll(randomWithPrefix(p), "-", "_")
}

func randomWithPrefix(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.NewString())
}
