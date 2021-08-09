package helpers

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Binding struct {
	serviceInstance ServiceInstance
	bindingName     string
	appInstance     AppInstance
}

func (b Binding) Credential() interface{} {
	out, _ := CF("app", "--guid", b.appInstance.name)
	guid := strings.TrimSpace(string(out))

	env, _ := CF("curl", fmt.Sprintf("/v3/apps/%s/env", guid))

	var receiver struct {
		Services map[string]interface{} `jsonry:"system_env_json.VCAP_SERVICES"`
	}
	err := jsonry.Unmarshal([]byte(env), &receiver)
	Expect(err).NotTo(HaveOccurred())

	Expect(receiver.Services).NotTo(BeEmpty())
	Expect(receiver.Services).To(HaveKey(b.serviceInstance.offering))
	bindings := receiver.Services[b.serviceInstance.offering]
	Expect(bindings).To(BeAssignableToTypeOf([]interface{}{}))
	Expect(bindings).NotTo(BeEmpty())

	for _, bnd := range bindings.([]interface{}) {
		if n, ok := bnd.(map[string]interface{})["name"]; ok && n == b.bindingName {
			Expect(bnd).To(HaveKey("credentials"))
			return bnd.(map[string]interface{})["credentials"]
		}
	}

	Fail(fmt.Sprintf("could not find data for binding: %s\n%+v", b.bindingName, bindings))
	return nil
}

func (b Binding) Unbind() {
	CF("unbind-service", b.appInstance.name, b.serviceInstance.name)
}
