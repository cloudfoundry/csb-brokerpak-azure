package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func AppBuild(source string) string {
	dir, err := os.MkdirTemp("", "")
	Expect(err).NotTo(HaveOccurred())

	name := path.Base(source)
	command := exec.Command("go", "build", "-o", fmt.Sprintf("%s/%s", dir, name))
	command.Dir = source
	command.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")

	session, err := Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, time.Minute).Should(Exit(0))

	err = os.WriteFile(path.Join(dir, "Procfile"), []byte(fmt.Sprintf("web: ./%s\n", name)), 0555)
	Expect(err).NotTo(HaveOccurred())

	return dir
}
