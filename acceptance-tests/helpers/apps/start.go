package apps

import (
	"time"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func (a *App) Start() {
	Start(a)
}

func Start(apps ...*App) {
	for _, app := range apps {
		session := cf.Start("start", app.Name)
		Eventually(session, 5*time.Minute).Should(gexec.Exit())
		checkSuccess(session.ExitCode(), app.Name)
	}
}
