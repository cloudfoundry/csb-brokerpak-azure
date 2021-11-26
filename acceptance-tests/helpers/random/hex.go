package random

import (
	"crypto/rand"
	"fmt"

	. "github.com/onsi/gomega"
)

func Hexadecimal(opts ...Option) string {
	length := cfg(append([]Option{WithMaxLength(20)}, opts...)).length
	buf := make([]byte, length/2)
	_, err := rand.Read(buf)
	Expect(err).NotTo(HaveOccurred())
	return fmt.Sprintf("%x", buf)
}
