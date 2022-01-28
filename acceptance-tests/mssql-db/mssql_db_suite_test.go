package mssql_db_test

import (
	"acceptancetests/helpers/environment"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMssqlDb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MssqlDb Suite")
}

var metadata environment.Metadata

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
})
