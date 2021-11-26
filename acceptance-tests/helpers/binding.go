package helpers

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/cf"
	"fmt"
	"strings"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Binding struct {
	serviceInstance ServiceInstance
	bindingName     string
	appInstance     apps.App
}

func (b Binding) Credential() interface{} {
	out, _ := cf.Run("app", "--guid", b.appInstance.Name)
	guid := strings.TrimSpace(string(out))

	env, _ := cf.Run("curl", fmt.Sprintf("/v3/apps/%s/env", guid))

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
	cf.Run("unbind-service", b.appInstance.Name, b.serviceInstance.name)
}
