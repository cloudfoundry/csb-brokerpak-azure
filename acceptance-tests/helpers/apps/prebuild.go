package apps

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func WithPreBuild(source string) Option {
	dir := newTmpDir()
	name := path.Base(source)
	command := exec.Command("go", "build", "-o", fmt.Sprintf("%s/%s", dir, name))
	command.Dir = source
	command.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")

	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, time.Minute).Should(gexec.Exit(0))

	err = os.WriteFile(path.Join(dir.path(), "Procfile"), []byte(fmt.Sprintf("web: ./%s\n", name)), 0555)
	Expect(err).NotTo(HaveOccurred())

	return WithOptions(
		WithBinaryBuildpack(),
		func(a *App) {
			a.dir = dir
		},
	)
}
