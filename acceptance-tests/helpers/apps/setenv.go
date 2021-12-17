package apps

import (
	"acceptancetests/helpers/cf"
	"encoding/json"

	. "github.com/onsi/gomega"
)

type EnvVar struct {
	Name  string
	Value interface{}
}

func (a *App) SetEnv(env ...EnvVar) {
	SetEnv(a.Name, env...)
}

func SetEnv(name string, env ...EnvVar) {
	for _, envVar := range env {
		switch v := envVar.Value.(type) {
		case string:
			if v == "" {
				cf.Run("unset-env", name, envVar.Name)
			} else {
				cf.Run("set-env", name, envVar.Name, v)
			}
		default:
			data, err := json.Marshal(v)
			Expect(err).NotTo(HaveOccurred())
			cf.Run("set-env", name, envVar.Name, string(data))
		}
	}
}
