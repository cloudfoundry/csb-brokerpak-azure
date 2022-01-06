package mssql_db_failover_group_test

import (
	"acceptancetests/helpers/environment"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMssqlDbFailoverGroup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MssqlDbFailoverGroup Suite")
}

var metadata environment.Metadata

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
})
