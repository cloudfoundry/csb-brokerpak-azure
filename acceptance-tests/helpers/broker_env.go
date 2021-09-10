package helpers

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

type EnvVar struct {
	Name  string
	Value interface{}
}

func SetBrokerEnv(envVars ...EnvVar) {
	const broker = "cloud-service-broker"

	for _, envVar := range envVars {
		switch v := envVar.Value.(type) {
		case string:
			if v == "" {
				CF("unset-env", broker, envVar.Name)
			} else {
				CF("set-env", broker, envVar.Name, v)
			}
		default:
			data, err := json.Marshal(v)
			Expect(err).NotTo(HaveOccurred())
			CF("set-env", broker, envVar.Name, string(data))
		}
	}

	CF("restart", broker)
}
