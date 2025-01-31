package brokers

import (
	"fmt"
	"os"

	"csbbrokerpakazure/acceptance-tests/helpers/testpath"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"

	"github.com/onsi/ginkgo/v2"
)

func (b *Broker) env() []apps.EnvVar {
	var result []apps.EnvVar

	for name, required := range map[string]bool{
		"ARM_SUBSCRIPTION_ID":        true,
		"ARM_TENANT_ID":              true,
		"ARM_CLIENT_ID":              true,
		"ARM_CLIENT_SECRET":          true,
		"GSB_BROKERPAK_BUILTIN_PATH": false,
		"GSB_PROVISION_DEFAULTS":     false,
		"CH_CRED_HUB_URL":            false,
		"CH_UAA_URL":                 false,
		"CH_UAA_CLIENT_NAME":         false,
		"CH_UAA_CLIENT_SECRET":       false,
		"CH_SKIP_SSL_VALIDATION":     false,
	} {
		val, ok := os.LookupEnv(name)
		switch {
		case ok:
			result = append(result, apps.EnvVar{Name: name, Value: val})
		case required:
			ginkgo.Fail(fmt.Sprintf("You must set the %s environment variable", name))
		}
	}

	result = append(result,
		apps.EnvVar{Name: "SECURITY_USER_NAME", Value: b.username},
		apps.EnvVar{Name: "SECURITY_USER_PASSWORD", Value: b.password},
		apps.EnvVar{Name: "DB_TLS", Value: "skip-verify"},
		apps.EnvVar{Name: "ENCRYPTION_ENABLED", Value: true},
		apps.EnvVar{Name: "ENCRYPTION_PASSWORDS", Value: b.secrets},
		apps.EnvVar{Name: "BROKERPAK_UPDATES_ENABLED", Value: true},
		apps.EnvVar{Name: "TERRAFORM_UPGRADES_ENABLED", Value: true},
		apps.EnvVar{Name: "CSB_DISABLE_TF_UPGRADE_PROVIDER_RENAMES", Value: true},
		apps.EnvVar{Name: "GSB_COMPATIBILITY_ENABLE_GCP_DEPRECATED_SERVICES", Value: true},
	)

	return append(result, b.envExtras...)
}

func (b *Broker) latestEnv() []apps.EnvVar {
	return readEnvrcServices(testpath.BrokerpakFile(".envrc"))
}
