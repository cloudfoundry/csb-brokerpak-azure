package apps

import (
	"acceptancetests/helpers/cf"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/jsonry"
)

func url(name string) string {
	env, _ := cf.Run("curl", fmt.Sprintf("/v3/apps/%s/env", guid(name)))
	var receiver struct {
		BrokerURL []string `jsonry:"application_env_json.VCAP_APPLICATION.application_uris[]"`
	}
	err := jsonry.Unmarshal([]byte(env), &receiver)
	Expect(err).NotTo(HaveOccurred())
	return fmt.Sprintf("http://%s", receiver.BrokerURL[0])
}

func guid(name string) string {
	out, _ := cf.Run("app", "--guid", name)
	return strings.TrimSpace(out)
}
