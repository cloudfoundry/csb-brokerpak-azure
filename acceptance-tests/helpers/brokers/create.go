package brokers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/cf"
	"csbbrokerpakazure/acceptance-tests/helpers/random"

	"github.com/onsi/gomega"
)

type Option func(broker *Broker)

func Create(opts ...Option) *Broker {
	broker := defaultConfig(opts...)

	brokerApp := apps.Push(
		apps.WithName(broker.Name),
		apps.WithDir(broker.dir),
		apps.WithManifest(newManifest(
			withName(broker.Name),
			withEnv(broker.env()...),
		)),
	)

	schemaName := strings.ReplaceAll(broker.Name, "-", "_")
	cf.Run("bind-service", broker.Name, "csb-sql", "-c", fmt.Sprintf(`{"schema":"%s"}`, schemaName))

	brokerApp.Start()

	cf.Run("create-service-broker", broker.Name, broker.username, broker.password, brokerApp.URL, "--space-scoped")

	broker.app = brokerApp
	return &broker
}

func WithOptions(opts ...Option) Option {
	return func(b *Broker) {
		for _, o := range opts {
			o(b)
		}
	}
}

func WithName(name string) Option {
	return func(b *Broker) {
		b.Name = name
	}
}

func WithPrefix(prefix string) Option {
	return func(b *Broker) {
		b.Name = random.Name(random.WithPrefix(prefix))
	}
}

func WithSourceDir(dir string) Option {
	return func(b *Broker) {
		gomega.Expect(filepath.Join(dir, "cloud-service-broker")).To(gomega.BeAnExistingFile())
		b.dir = dir
	}
}
func WithConfig(config map[string]interface{}, dir string) {

	bytes, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	os.WriteFile(fmt.Sprintf("%s/config.yml", dir), bytes, 0666)

}
func WithEnv(env ...apps.EnvVar) Option {

	return func(b *Broker) {
		b.envExtras = append(b.envExtras, env...)
	}
}

func WithUsername(username string) Option {
	return func(b *Broker) {
		b.username = username
	}
}

func WithPassword(password string) Option {
	return func(b *Broker) {
		b.password = password
	}
}

func defaultConfig(opts ...Option) (broker Broker) {
	defaults := []Option{
		WithName(random.Name(random.WithPrefix("broker"))),
		WithSourceDir(defaultSourceDir()),
		WithUsername(random.Name()),
		WithPassword(random.Password()),
		WithEncryptionSecret(random.Password()),
	}
	WithOptions(append(defaults, opts...)...)(&broker)
	return broker
}

func defaultSourceDir() string {
	for _, d := range []string{"..", "../.."} {
		p := fmt.Sprintf("%s/%s", d, "cf-manifest.yml")
		_, err := os.Stat(p)
		if err == nil {
			return d
		}
	}

	panic("could not find source for broker app")
}
