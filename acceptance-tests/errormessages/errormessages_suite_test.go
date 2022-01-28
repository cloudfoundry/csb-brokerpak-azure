package errormessages_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestErrormessages(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Error Messages Suite")
}
