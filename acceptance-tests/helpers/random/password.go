package random

import (
	"crypto/rand"
	"regexp"
	"strings"

	. "github.com/onsi/gomega"
)

func Password(opts ...Option) string {
	length := cfg(append([]Option{WithMaxLength(24)}, opts...)).length
	var s strings.Builder
	for s.Len() < length {
		buf := make([]byte, 1)
		_, err := rand.Read(buf)
		Expect(err).NotTo(HaveOccurred())
		if regexp.MustCompile(`[-~_.a-zA-Z0-9]`).MatchString(string(buf)) {
			s.WriteString(string(buf))
		}
	}

	return s.String()
}
