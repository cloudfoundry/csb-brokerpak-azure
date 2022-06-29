// Package testhelpers contains some helpers shared between different test packages
package testhelpers

import (
	"fmt"
	"net"

	"github.com/onsi/gomega"
	"github.com/pborman/uuid"
)

func FreePort() int {
	listener, err := net.Listen("tcp", "localhost:0")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func RandomPassword() string {
	return fmt.Sprintf("AaZz09~._%s", uuid.New())
}
