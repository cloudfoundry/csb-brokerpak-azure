// Package lookupplan is used for looking up plan information from services
package lookupplan

import (
	"fmt"
	"strings"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// LookupByID looks up a plan by broker ID. There were historical bugs where duplicate plan ID were used, so we take in
// the service offering name too
func LookupByID(id, serviceOfferingName, serviceBrokerName string) string {
	data, _ := cf.Run("curl", fmt.Sprintf("/v3/service_plans?service_broker_names=%s&service_offering_names=%s", serviceBrokerName, serviceOfferingName))

	var receiver struct {
		Resources []struct {
			Name string `json:"name"`
			ID   string `jsonry:"broker_catalog.id"`
		} `json:"resources"`
	}
	Expect(jsonry.Unmarshal([]byte(data), &receiver)).To(Succeed())

	var matches []string
	for _, e := range receiver.Resources {
		if e.ID == id {
			matches = append(matches, e.Name)
		}
	}

	switch len(matches) {
	case 0:
		Fail(fmt.Sprintf("could not find match for plan ID: %s", id))
	case 1:
		// ok
	default:
		Fail(fmt.Sprintf("too many matches for plan ID %q: %s", id, strings.Join(matches, ", ")))
	}
	return matches[0]
}
