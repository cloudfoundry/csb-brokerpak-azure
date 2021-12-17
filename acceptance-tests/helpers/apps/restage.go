package apps

import (
	"acceptancetests/helpers/cf"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func (a *App) Restage() {
	Restage(a)
}

func Restage(apps ...*App) {
	for _, app := range apps {
		session := cf.Start("restage", app.Name)
		Eventually(session, 5*time.Minute).Should(gexec.Exit())
		checkSuccess(session.ExitCode(), app.Name)
	}
}
