package brokers

import (
	"fmt"
	"os"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/gomega"
)

var defaultBrokerName string

func DefaultBrokerName() string {
	if defaultBrokerName != "" {
		return defaultBrokerName
	}

	var receiver struct {
		Names []string `jsonry:"resources.name"`
	}
	out, _ := cf.Run("curl", "/v3/service_brokers")
	Expect(jsonry.Unmarshal([]byte(out), &receiver)).NotTo(HaveOccurred())

	username := os.Getenv("USER")
	for _, n := range receiver.Names {
		if n == "broker-cf-test" || n == "cloud-service-broker-azure" {
			defaultBrokerName = n
			return n
		}

		if username != "" && n == fmt.Sprintf("csb-%s", username) {
			defaultBrokerName = n
			return n
		}
	}

	panic("could not determine default broker name")
}
