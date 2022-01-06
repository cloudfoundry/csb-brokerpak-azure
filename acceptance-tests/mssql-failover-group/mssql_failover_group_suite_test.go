package mssql_failover_group_test

import (
	"acceptancetests/helpers/environment"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMssqlFailoverGroup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MssqlFailoverGroup Suite")
}

var metadata environment.Metadata

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
})
