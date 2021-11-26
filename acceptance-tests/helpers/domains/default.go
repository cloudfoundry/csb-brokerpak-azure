package domains

import (
	"acceptancetests/helpers/cf"
	"regexp"
	"strings"

	"github.com/onsi/gomega"
)

var defaultDomain string
var regex = regexp.MustCompile(`^(\S+)\s+shared\s+(http)?\s*$`)

func Default() string {
	if defaultDomain == "" {
		output, _ := cf.Run("domains")
		for _, line := range strings.Split(output, "\n") {
			matches := regex.FindStringSubmatch(line)
			if len(matches) > 0 {
				defaultDomain = matches[1]
			}
		}
	}

	gomega.Expect(defaultDomain).NotTo(gomega.BeEmpty())
	return defaultDomain
}
