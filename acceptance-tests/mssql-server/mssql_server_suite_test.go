package mssql_server_test

import (
	"acceptancetests/helpers/environment"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMssqlServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MssqlServer Suite")
}

var metadata environment.Metadata

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
})
