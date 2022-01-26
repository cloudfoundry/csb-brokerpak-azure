package random

import (
	"crypto/rand"
	"regexp"
	"strings"

	. "github.com/onsi/gomega"
)

var (
	firstLetter      = regexp.MustCompile(`^[a-zA-Z]$`)
	subsequentLetter = regexp.MustCompile(`^[~_.a-zA-Z0-9]$`)
)

func Password(opts ...Option) string {
	length := cfg(append([]Option{WithMaxLength(24)}, opts...)).length
	var s strings.Builder

	s.WriteByte(byteMatching(firstLetter))

	for s.Len() < length {
		s.WriteByte(byteMatching(subsequentLetter))
	}

	return s.String()
}

func byteMatching(re *regexp.Regexp) byte {
	buf := make([]byte, 1)
	for {
		_, err := rand.Read(buf)
		Expect(err).NotTo(HaveOccurred())
		if re.MatchString(string(buf)) {
			return buf[0]
		}
	}
}
