package brokers

import "acceptancetests/helpers/apps"

type Broker struct {
	Name      string
	username  string
	password  string
	secrets   []EncryptionSecret
	dir       string
	envExtras []apps.EnvVar
	app       *apps.App
}
