package helpers

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func GetBindingCredential(appName, serviceName, bindingName string) interface{} {
	out, _ := CF("app", "--guid", appName)
	guid := strings.TrimSpace(string(out))

	env, _ := CF("curl", fmt.Sprintf("/v3/apps/%s/env", guid))

	var receiver struct {
		Services map[string]interface{} `jsonry:"system_env_json.VCAP_SERVICES"`
	}
	err := jsonry.Unmarshal([]byte(env), &receiver)
	Expect(err).NotTo(HaveOccurred())

	Expect(receiver.Services).NotTo(BeEmpty())
	Expect(receiver.Services).To(HaveKey(serviceName))
	bindings := receiver.Services[serviceName]
	Expect(bindings).To(BeAssignableToTypeOf([]interface{}{}))
	Expect(bindings).NotTo(BeEmpty())

	for _, b := range bindings.([]interface{}) {
		if n, ok := b.(map[string]interface{})["name"]; ok && n == bindingName {
			Expect(b).To(HaveKey("credentials"))
			return b.(map[string]interface{})["credentials"]
		}
	}

	Fail(fmt.Sprintf("could not find data for binding: %s\n%+v", bindingName, bindings))
	return nil
}
