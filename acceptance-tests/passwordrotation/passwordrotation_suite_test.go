package passwordrotation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKeyrotation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Passwordrotation Suite")
}
