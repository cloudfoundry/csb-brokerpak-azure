package acceptance_test

import (
	"acceptancetests/helpers/environment"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAcceptanceTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Tests Suite")
}

var metadata environment.Metadata

var _ = BeforeSuite(func() {
	metadata = environment.ReadMetadata()
})
