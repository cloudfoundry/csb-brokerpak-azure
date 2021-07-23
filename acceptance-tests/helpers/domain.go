package helpers

import (
	"regexp"
	"strings"

	. "github.com/onsi/gomega"
)

var defaultDomain string
var regex = regexp.MustCompile(`^(\S+)\s+shared\s+(http)?\s*$`)

func DefaultSharedDomain() string {
	if defaultDomain == "" {
		output, _ := CF("domains")
		for _, line := range strings.Split(output, "\n") {
			matches := regex.FindStringSubmatch(line)
			if len(matches) > 0 {
				defaultDomain = matches[1]
			}
		}
	}

	Expect(defaultDomain).NotTo(BeEmpty())
	return defaultDomain
}
