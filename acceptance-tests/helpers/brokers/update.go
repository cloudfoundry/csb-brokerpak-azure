package brokers

import (
	"slices"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/cf"
)

func (b *Broker) UpgradeBroker(dir string, env ...apps.EnvVar) {
	b.envExtras = slices.Concat(b.envExtras, b.latestEnv(), env)

	b.app.Push(
		apps.WithName(b.Name),
		apps.WithDir(dir),
		apps.WithStartedState(),
		apps.WithManifest(newManifest(
			withName(b.Name),
			withEnv(b.env()...),
		)),
	)

	cf.Run("update-service-broker", b.Name, b.username, b.password, b.app.URL)
}

func (b *Broker) UpdateEnv(env ...apps.EnvVar) {
	WithEnv(env...)(b)
	b.app.SetEnv(env...)
	b.app.Restart()

	cf.Run("update-service-broker", b.Name, b.username, b.password, b.app.URL)
}

func (b *Broker) UpdateConfig(config map[string]interface{}) {
	b.app.Push(
		apps.WithName(b.Name),
		apps.WithYAMLFile("config.yml", config),
		apps.WithManifest(newManifest(
			withName(b.Name),
			withEnv(b.env()...),
			withCustomStartCommand("./cloud-service-broker --config config.yml serve"),
		)),
	)

	b.app.CleanFileFromAppDir("config.yml")

	b.app.Start()
}

func (b *Broker) UpdateEncryptionSecrets(secrets ...EncryptionSecret) {
	WithEncryptionSecrets(secrets...)
	b.app.SetEnv(b.env()...)

	cf.Run("update-service-broker", b.Name, b.username, b.password, b.app.URL)
}
