package bindings

import (
	"fmt"
	"strings"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func (b *Binding) Credential() any {
	out, _ := cf.Run("app", "--guid", b.appName)
	env, _ := cf.Run("curl", fmt.Sprintf("/v3/apps/%s/env", strings.TrimSpace(out)))

	var receiver struct {
		Services map[string]any `jsonry:"system_env_json.VCAP_SERVICES"`
	}
	Expect(jsonry.Unmarshal([]byte(env), &receiver)).NotTo(HaveOccurred())

	for _, bindings := range receiver.Services {
		Expect(bindings).To(BeAssignableToTypeOf([]any{}))
		for _, bnd := range bindings.([]any) {
			if n, ok := bnd.(map[string]any)["name"]; ok && n == b.name {
				Expect(bnd).To(HaveKey("credentials"))
				return bnd.(map[string]any)["credentials"]
			}
		}
	}

	Fail(fmt.Sprintf("could not find data for binding: %q\n%+v", b.name, receiver.Services))
	return nil
}
