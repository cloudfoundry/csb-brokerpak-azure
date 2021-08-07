package withoutcredhub_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWithoutCredHub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Without CredHub")
}
