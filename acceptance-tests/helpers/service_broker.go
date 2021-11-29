package helpers

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	brokerUsername            = "brokeruser"
	brokerPassword            = "brokeruserpassword"
	broker                    = "cloud-service-broker"
	encryptionEnabledEnvVar   = "ENCRYPTION_ENABLED"
	encryptionPasswordsEnvVar = "ENCRYPTION_PASSWORDS"
	cfOperationWaitTime       = 20 * time.Minute
)

type ServiceBroker struct {
	Name string
}

type Option func(*config)

type config struct {
	name string
	env  []apps.EnvVar
	dir  string
}

func CreateBroker(opts ...Option) ServiceBroker {
	cfg := config{
		name: random.Name(random.WithPrefix("csb")),
		env:  nil,
		dir:  "../..",
	}

	for _, o := range opts {
		o(&cfg)
	}

	brokerApp := apps.Push(
		apps.WithName(cfg.name),
		apps.WithDir(cfg.dir),
		apps.WithManifest(fmt.Sprintf("%s/cf-manifest.yml", cfg.dir)),
		apps.WithVariable("app", cfg.name),
	)
	setEnvVars(brokerApp, cfg.env...)

	schemaName := strings.ReplaceAll(cfg.name, "-", "_")
	cf.Run("bind-service", cfg.name, "csb-sql", "-c", fmt.Sprintf(`{"schema":"%s"}`, schemaName))

	brokerApp.Start()

	session := cf.Start("create-service-broker", cfg.name, brokerUsername, brokerPassword, brokerApp.URL, "--space-scoped")
	waitForBrokerOperation(session, cfg.name)

	return ServiceBroker{
		Name: cfg.name,
	}
}

func BrokerWithPrefix(prefix string) Option {
	return func(c *config) {
		c.name = random.Name(random.WithPrefix(prefix))
	}
}

func BrokerWithEnv(env ...apps.EnvVar) Option {
	return func(c *config) {
		c.env = env
	}
}

func BrokerFromDir(dir string) Option {
	return func(c *config) {
		c.dir = dir
	}
}

func DefaultBroker() ServiceBroker {
	return ServiceBroker{
		Name: "broker-cf-test",
	}
}

func (b ServiceBroker) Update(brokerDir string) {
	brokerApp := apps.Push(apps.WithName(b.Name), apps.WithDir(brokerDir))

	session := cf.Start("update-service-broker", b.Name, brokerUsername, brokerPassword, brokerApp.URL)
	waitForBrokerOperation(session, b.Name)
}

func (b ServiceBroker) Delete() {
	session := cf.Start("delete-service-broker", b.Name, "-f")
	waitForBrokerOperation(session, b.Name)

	session = cf.Start("delete", b.Name, "-f")
	waitForAppDelete(session, b.Name)
}

func setEnvVars(app apps.App, extra ...apps.EnvVar) {
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
		apps.EnvVar{Name: "SECURITY_USER_NAME", Value: brokerUsername},
		apps.EnvVar{Name: "SECURITY_USER_PASSWORD", Value: brokerPassword},
		apps.EnvVar{Name: "DB_TLS", Value: "skip-verify"},
		apps.EnvVar{Name: "ENCRYPTION_ENABLED", Value: true},
		apps.EnvVar{Name: "ENCRYPTION_PASSWORDS", Value: `[{"password": {"secret":"superSecretP@SSw0Rd1234!"},"label":"first-encryption","primary":true}]`},
	)

	envVars = append(envVars, extra...)

	app.SetEnv(envVars...)
}

func optionalEnvVar(envVars ...string) []apps.EnvVar {
	var toSet []apps.EnvVar
	for _, envVarName := range envVars {
		value, set := os.LookupEnv(envVarName)
		if set {
			toSet = append(toSet, apps.EnvVar{Name: envVarName, Value: value})
		}
	}
	return toSet
}

func requiredEnvVar(envVars ...string) []apps.EnvVar {
	var toSet []apps.EnvVar
	for _, envVarName := range envVars {
		value := os.Getenv(envVarName)
		Expect(value).NotTo(BeEmpty(), fmt.Sprintf("You must set the %s environment variable", envVarName))
		toSet = append(toSet, apps.EnvVar{Name: envVarName, Value: value})
	}
	return toSet
}

func waitForBrokerOperation(session *Session, name string) {
	Eventually(session, 5*time.Minute).Should(Exit())
	Expect(session.ExitCode()).To(BeZero())
}

func RestartBroker(broker string) {
	cf.Run("restart", broker)
}

func SetBrokerEnvAndRestart(envVars ...apps.EnvVar) {
	for _, envVar := range envVars {
		switch v := envVar.Value.(type) {
		case string:
			if v == "" {
				cf.Run("unset-env", broker, envVar.Name)
			} else {
				cf.Run("set-env", broker, envVar.Name, v)
			}
		default:
			data, err := json.Marshal(v)
			Expect(err).NotTo(HaveOccurred())
			cf.Run("set-env", broker, envVar.Name, string(data))
		}
	}

	cf.Run("restart", broker)
}

func GetBrokerEncryptionEnv(broker string) BrokerEnvVars {
	out, _ := cf.Run("app", "--guid", broker)
	guid := strings.TrimSpace(string(out))

	env, _ := cf.Run("curl", fmt.Sprintf("/v3/apps/%s/environment_variables", guid))

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
		EncryptionEnabled:   receiver.Var[encryptionEnabledEnvVar] == "true",
	}
}

func SetBrokerEncryptionEnv(brokerName string, brokerEnvVars BrokerEnvVars) {
	envVars := []apps.EnvVar{
		{
			Name:  encryptionEnabledEnvVar,
			Value: brokerEnvVars.EncryptionEnabled,
		},
		{
			Name:  encryptionPasswordsEnvVar,
			Value: brokerEnvVars.EncryptionPasswords,
		},
	}
	apps.SetEnv(brokerName, envVars...)
	session := cf.Start("restart", brokerName)
	waitForAppPush(session, brokerName)
}

type BrokerEnvVars struct {
	EncryptionPasswords EncryptionPasswords
	EncryptionEnabled   bool
}

type EncryptionPasswords []EncryptionPassword

type EncryptionPassword struct {
	Password Password `json:"password"`
	Label    string   `json:"label"`
	Primary  bool     `json:"primary"`
}

type Password struct {
	Secret string `json:"secret"`
}

func waitForAppPush(session *Session, name string) apps.App {
	Eventually(session, cfOperationWaitTime).Should(Exit())

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to push app. Getting logs...")
		cf.Run("logs", name, "--recent")
		Fail("App failed to push")
	}

	return apps.App{Name: name}
}

func waitForAppDelete(session *Session, name string) apps.App {
	Eventually(session, cfOperationWaitTime).Should(Exit())

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to delete app. Getting logs...")
		cf.Run("logs", name, "--recent")
		Fail("App failed to delete")
	}

	return apps.App{Name: name}
}
