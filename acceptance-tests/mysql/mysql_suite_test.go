package mysql_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRedis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MySQL Suite")
}
