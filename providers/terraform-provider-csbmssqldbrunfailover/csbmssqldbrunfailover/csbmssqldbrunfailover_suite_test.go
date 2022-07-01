package csbmssqldbrunfailover

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCSBSQLServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSB MSSQL Db Run Failover Terraform Provider Suite")
}
