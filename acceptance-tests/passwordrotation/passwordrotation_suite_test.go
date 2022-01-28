package passwordrotation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKeyRotation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Password Rotation Suite")
}
