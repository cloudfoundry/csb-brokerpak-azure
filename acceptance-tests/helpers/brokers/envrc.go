package brokers

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"

	. "github.com/onsi/gomega"
)

var serviceMatcher = regexp.MustCompile(`^\s*export\s+(GSB_SERVICE_[\w_]+?)='(.*)'\s*$`)

func readEnvrcServices(path string) (result []apps.EnvVar) {
	data, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())

	for _, line := range strings.Split(string(data), "\n") {
		m := serviceMatcher.FindStringSubmatch(line)
		const expectedNumberOfMatches = 3
		if len(m) != expectedNumberOfMatches {
			continue
		}
		name, rawValue := m[1], m[2]

		var r any
		Expect(json.Unmarshal([]byte(rawValue), &r)).To(Succeed(), func() string {
			return fmt.Sprintf("JSON parsing error %q for service %q: %s", err, name, rawValue)
		})
		tidyValue, err := json.Marshal(r)
		Expect(err).NotTo(HaveOccurred())

		result = append(result, apps.EnvVar{Name: name, Value: string(tidyValue)})
	}

	return result
}
