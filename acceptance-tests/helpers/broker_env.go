package helpers

import (
	"code.cloudfoundry.org/jsonry"
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
)

type EnvVar struct {
	Name  string
	Value interface{}
}

const broker = "cloud-service-broker"
const encryptionEnabled = "ENCRYPTION_ENABLED"
const encryptionPasswords = "ENCRYPTION_PASSWORDS"

func SetBrokerEnv(envVars ...EnvVar) {
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

func GetBrokerEncryptionEnv() BrokerEnvVars {
	out, _ := CF("app", "--guid", broker)
	guid := strings.TrimSpace(string(out))

	env, _ := CF("curl", fmt.Sprintf("/v3/apps/%s/environment_variables", guid))

	var receiver struct {
		Var map[string]string `jsonry:"var"`
	}
	err := jsonry.Unmarshal([]byte(env), &receiver)
	Expect(err).NotTo(HaveOccurred())

	var encryptionPasswords EncryptionPasswords
	err = json.Unmarshal([]byte(receiver.Var[encryptionPasswords]), &encryptionPasswords)
	Expect(err).NotTo(HaveOccurred())

	return BrokerEnvVars{
		EncryptionPasswords: encryptionPasswords,
		EncryptionEnabled:  receiver.Var[encryptionEnabled] == "true",
	}
}

func SetBrokerEncryptionEnv(brokerEnvVars BrokerEnvVars) {
	envVars := []EnvVar{
		{
			Name: encryptionEnabled,
			Value: brokerEnvVars.EncryptionEnabled,
		},
		{
			Name: encryptionPasswords,
			Value: brokerEnvVars.EncryptionPasswords,
		},
	}
	SetBrokerEnv(envVars...)
}

type BrokerEnvVars  struct {
	EncryptionPasswords EncryptionPasswords
	EncryptionEnabled bool
}

type EncryptionPasswords []EncryptionPassword

type EncryptionPassword struct {
	Password Password `json:"password"`
	Label string `json:"label"`
	Primary bool `json:"primary"`
}

type Password struct {
	Secret string `json:"secret"`
}