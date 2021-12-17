package brokers

import "code.cloudfoundry.org/jsonry"

type EncryptionSecret struct {
	Password string `jsonry:"password.secret"`
	Label    string `json:"label"`
	Primary  bool   `json:"primary"`
}

func (e *EncryptionSecret) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(e)
}

func WithEncryptionSecret(password string) Option {
	return func(b *Broker) {
		b.secrets = append(b.secrets, EncryptionSecret{
			Password: password,
			Label:    "default",
			Primary:  true,
		})
	}
}

func WithEncryptionSecrets(secrets ...EncryptionSecret) Option {
	return func(b *Broker) {
		b.secrets = secrets
	}
}
