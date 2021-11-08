package helpers

import (
	"code.cloudfoundry.org/jsonry"
	"encoding/json"
	"fmt"
	. "github.com/onsi/gomega"
	"os"
	"strings"
	"time"

	. "github.com/onsi/gomega/gexec"
)

const brokerUsername = "brokeruser"
const brokerPassword = "brokeruserpassword"
const broker = "cloud-service-broker"
const encryptionEnabledEnvVar = "ENCRYPTION_ENABLED"
const encryptionPasswordsEnvVar = "ENCRYPTION_PASSWORDS"

type ServiceBroker struct {
	Name          string
	mySqlInstance ServiceInstance
}

func DefaultBroker() ServiceBroker {
	return ServiceBroker{
		Name: "broker-cf-test",
	}
}

func PushAndStartBroker(brokerName, brokerDir string) ServiceBroker {
	brokerApp := pushNoStartServiceBroker(brokerName, brokerDir)
	setEnvVars(brokerName)

	mySqlInstance := CreateService("p.mysql", "db-small")
	CF("bind-service", brokerName, mySqlInstance.name)

	session := StartCF("restart", brokerName)
	waitForAppPush(session, brokerName)

	brokerURL := getBrokerAppURL(brokerApp)
	session = StartCF("create-service-broker", brokerName, brokerUsername, brokerPassword, "https://"+brokerURL, "--space-scoped")
	waitForBrokerOperation(session, brokerName)

	return ServiceBroker{
		Name:          brokerName,
		mySqlInstance: mySqlInstance,
	}
}

func (b ServiceBroker) Update(brokerDir string) {
	brokerApp := pushServiceBroker(b.Name, brokerDir)

	brokerURL := getBrokerAppURL(brokerApp)
	session := StartCF("update-service-broker", b.Name, brokerUsername, brokerPassword, "https://"+brokerURL)
	waitForBrokerOperation(session, b.Name)
}

func (b ServiceBroker) Delete() {
	session := StartCF("delete-service-broker", b.Name, "-f")
	waitForBrokerOperation(session, b.Name)

	session = StartCF("delete", b.Name, "-f")
	waitForAppDelete(session, b.Name)

	b.mySqlInstance.Delete()
}

func setEnvVars(brokerName string) {
	envVars := requiredEnvVar(
		"ARM_SUBSCRIPTION_ID",
		"ARM_TENANT_ID",
		"ARM_CLIENT_ID",
		"ARM_CLIENT_SECRET",
	)

	envVars = append(envVars, optionalEnvVar(
		"GSB_BROKERPAK_BUILTIN_PATH",
		"GSB_PROVISION_DEFAULTS",
		"CH_CRED_HUB_URL",
		"CH_UAA_URL",
		"CH_UAA_CLIENT_NAME",
		"CH_UAA_CLIENT_SECRET",
		"CH_SKIP_SSL_VALIDATION",
	)...)

	envVars = append(envVars,
		EnvVar{Name: "SECURITY_USER_NAME", Value: brokerUsername},
		EnvVar{Name: "SECURITY_USER_PASSWORD", Value: brokerPassword},
		EnvVar{Name: "DB_TLS", Value: "skip-verify"},
		EnvVar{Name: "ENCRYPTION_ENABLED", Value: true},
		EnvVar{Name: "ENCRYPTION_PASSWORDS", Value: `[{"password": {"secret":"superSecretP@SSw0Rd1234!"},"label":"first-encryption","primary":true}]`},
	)

	SetBrokerEnv(brokerName, envVars...)
}

func optionalEnvVar(envVars ...string) []EnvVar {
	var toSet []EnvVar
	for _, envVarName := range envVars {
		value, set := os.LookupEnv(envVarName)
		if set {
			toSet = append(toSet, EnvVar{Name: envVarName, Value: value})
		}
	}
	return toSet
}

func requiredEnvVar(envVars ...string) []EnvVar {
	var toSet []EnvVar
	for _, envVarName := range envVars {
		value := os.Getenv(envVarName)
		Expect(value).NotTo(BeEmpty(), fmt.Sprintf("You must set the %s environment variable", envVarName))
		toSet = append(toSet, EnvVar{Name: envVarName, Value: value})
	}
	return toSet
}

func getBrokerAppURL(brokerApp AppInstance) string {
	out, _ := CF("app", "--guid", brokerApp.name)
	guid := strings.TrimSpace(out)
	env, _ := CF("curl", fmt.Sprintf("/v3/apps/%s/env", guid))

	var receiver struct {
		BrokerURL []string `jsonry:"application_env_json.VCAP_APPLICATION.application_uris[]"`
	}
	err := jsonry.Unmarshal([]byte(env), &receiver)
	Expect(err).NotTo(HaveOccurred())
	return receiver.BrokerURL[0]
}

func pushNoStartServiceBroker(brokerName, brokerDir string) AppInstance {
	session := StartCF("push", brokerName, "--no-start", "-p", brokerDir, "-f", fmt.Sprintf("%s/cf-manifest.yml", brokerDir), "--var", fmt.Sprintf("app=%s", brokerName))
	return waitForAppPush(session, brokerName)
}

func pushServiceBroker(brokerName, brokerDir string) AppInstance {
	session := StartCF("push", brokerName, "-p", brokerDir, "-f", fmt.Sprintf("%s/cf-manifest.yml", brokerDir), "--var", fmt.Sprintf("app=%s", brokerName))
	return waitForAppPush(session, brokerName)
}

func waitForBrokerOperation(session *Session, name string) {
	Eventually(session, 5*time.Minute).Should(Exit())
	Expect(session.ExitCode()).To(BeZero())
}

type EnvVar struct {
	Name  string
	Value interface{}
}

func SetBrokerEnv(brokerName string, envVars ...EnvVar) {
	for _, envVar := range envVars {
		switch v := envVar.Value.(type) {
		case string:
			if v == "" {
				CF("unset-env", brokerName, envVar.Name)
			} else {
				CF("set-env", brokerName, envVar.Name, v)
			}
		default:
			data, err := json.Marshal(v)
			Expect(err).NotTo(HaveOccurred())
			CF("set-env", brokerName, envVar.Name, string(data))
		}
	}
}

func RestartBroker(broker string) {
	CF("restart", broker)
}

func SetBrokerEnvAndRestart(envVars ...EnvVar) {
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

func GetBrokerEncryptionEnv(broker string) BrokerEnvVars {
	out, _ := CF("app", "--guid", broker)
	guid := strings.TrimSpace(string(out))

	env, _ := CF("curl", fmt.Sprintf("/v3/apps/%s/environment_variables", guid))

	var receiver struct {
		Var map[string]string `jsonry:"var"`
	}
	err := jsonry.Unmarshal([]byte(env), &receiver)
	Expect(err).NotTo(HaveOccurred())

	var encryptionPasswords EncryptionPasswords
	err = json.Unmarshal([]byte(receiver.Var[encryptionPasswordsEnvVar]), &encryptionPasswords)
	Expect(err).NotTo(HaveOccurred())

	return BrokerEnvVars{
		EncryptionPasswords: encryptionPasswords,
		EncryptionEnabled:  receiver.Var[encryptionEnabledEnvVar] == "true",
	}
}

func SetBrokerEncryptionEnv(brokerName string, brokerEnvVars BrokerEnvVars) {
	envVars := []EnvVar{
		{
			Name: encryptionEnabledEnvVar,
			Value: brokerEnvVars.EncryptionEnabled,
		},
		{
			Name: encryptionPasswordsEnvVar,
			Value: brokerEnvVars.EncryptionPasswords,
		},
	}
	SetBrokerEnv(brokerName, envVars...)
	session := StartCF("restart", brokerName)
	waitForAppPush(session, brokerName)
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