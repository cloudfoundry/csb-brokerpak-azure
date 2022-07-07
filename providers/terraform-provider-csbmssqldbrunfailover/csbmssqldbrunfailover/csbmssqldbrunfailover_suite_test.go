package csbmssqldbrunfailover_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCSBSQLDBRunFailoverServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSB MSSQL Db Run Failover Terraform Provider Suite")
}
