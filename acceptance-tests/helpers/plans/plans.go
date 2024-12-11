// Package plans provides plan helper functions
package plans

import (
	"encoding/json"
	"fmt"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
)

func ExistsAndAvailable(planName, offeringName, brokerName string) bool {
	plansJSON, err := cf.Run("curl", fmt.Sprintf("v3/service_plans?names=%s&service_broker_names=%s&service_offering_names=%s&available=true", planName, brokerName, offeringName))
	Expect(err).To(BeEmpty())

	type plan struct {
		GUID string `json:"guid"`
	}

	var receiver struct {
		Plans []plan `json:"resources"`
	}

	Expect(json.Unmarshal([]byte(plansJSON), &receiver)).NotTo(HaveOccurred())
	return len(receiver.Plans) > 0
}
