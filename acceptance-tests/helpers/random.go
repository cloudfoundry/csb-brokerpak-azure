package helpers

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"

	"github.com/Pallinder/go-randomdata"
	. "github.com/onsi/gomega"
)

func RandomName(prefixes ...string) string {
	return strings.Join(append(prefixes, randomdata.Adjective(), randomdata.Noun()), "-")
}

func RandomShortName() string {
	return randomdata.Noun()
}

func RandomHex() string {
	const numBytes = 10
	buf := make([]byte, numBytes)
	_, err := rand.Read(buf)
	Expect(err).NotTo(HaveOccurred())
	return fmt.Sprintf("%x", buf)
}

func RandomPassword() string {
	var s strings.Builder
	for s.Len() < 24 {
		buf := make([]byte, 1)
		_, err := rand.Read(buf)
		Expect(err).NotTo(HaveOccurred())
		if regexp.MustCompile(`[-~_.a-zA-Z0-9]`).MatchString(string(buf)) {
			s.WriteString(string(buf))
		}
	}

	return s.String()
}
