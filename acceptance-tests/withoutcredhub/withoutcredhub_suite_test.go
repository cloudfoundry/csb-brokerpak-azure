package withoutcredhub_test

import (
	"acceptancetests/helpers"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWithoutCredHub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Without CredHub")
}

var _ = BeforeSuite(func() {
	helpers.SetBrokerEnv(helpers.EnvVar{
		Name:  "CH_CRED_HUB_URL",
		Value: "",
	})
})

var _ = AfterSuite(func() {
	helpers.SetBrokerEnv(helpers.EnvVar{
		Name:  "CH_CRED_HUB_URL",
		Value: "https://credhub.service.cf.internal:8844",
	})
})
