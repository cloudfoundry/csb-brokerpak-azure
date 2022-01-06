package bindings

import (
	"acceptancetests/helpers/cf"
	"fmt"
	"strings"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func (b *Binding) Credential() interface{} {
	out, _ := cf.Run("app", "--guid", b.appName)
	env, _ := cf.Run("curl", fmt.Sprintf("/v3/apps/%s/env", strings.TrimSpace(out)))

	var receiver struct {
		Services map[string]interface{} `jsonry:"system_env_json.VCAP_SERVICES"`
	}
	Expect(jsonry.Unmarshal([]byte(env), &receiver)).NotTo(HaveOccurred())

	for _, bindings := range receiver.Services {
		Expect(bindings).To(BeAssignableToTypeOf([]interface{}{}))
		for _, bnd := range bindings.([]interface{}) {
			if n, ok := bnd.(map[string]interface{})["name"]; ok && n == b.name {
				Expect(bnd).To(HaveKey("credentials"))
				return bnd.(map[string]interface{})["credentials"]
			}
		}
	}

	Fail(fmt.Sprintf("could not find data for binding: %q\n%+v", b.name, receiver.Services))
	return nil
}
