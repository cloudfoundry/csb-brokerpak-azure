package apps

import (
	"encoding/json"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
)

type EnvVar struct {
	Name  string
	Value any
}

func (e EnvVar) ValueString() string {
	switch v := e.Value.(type) {
	case string:
		return v
	default:
		data, err := json.Marshal(v)
		Expect(err).NotTo(HaveOccurred())
		return string(data)
	}
}

func (a *App) SetEnv(env ...EnvVar) {
	SetEnv(a.Name, env...)
}

func SetEnv(name string, env ...EnvVar) {
	for _, envVar := range env {
		v := envVar.ValueString()
		if v == "" {
			cf.Run("unset-env", name, envVar.Name)
		} else {
			cf.Run("set-env", name, envVar.Name, v)
		}
	}
}
