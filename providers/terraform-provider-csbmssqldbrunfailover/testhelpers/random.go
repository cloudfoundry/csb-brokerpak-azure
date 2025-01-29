package testhelpers

import (
	"crypto/rand"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func RandomName(prefix string) string {
	GinkgoHelper()

	return fmt.Sprintf("%s-%s", prefix, RandomHex())
}

func RandomHex() string {
	GinkgoHelper()

	const length = 20
	buf := make([]byte, length/2)
	_, err := rand.Read(buf)
	Expect(err).NotTo(HaveOccurred())
	return fmt.Sprintf("%x", buf)
}
