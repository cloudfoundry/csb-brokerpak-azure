package helpers

import (
	"acceptancetests/apps"
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const cfOperationWaitTime = 20 * time.Minute

type AppInstance struct {
	name string
}

func AppPushUnstarted(app apps.AppCode) AppInstance {
	switch app {
	case apps.Cosmos, apps.Storage:
		return appPushUnstartedNoBuildpack(app)
	default:
		return appBuildAndPushUnstartedBinaryBuildpack(app)
	}
}

func appPushUnstartedNoBuildpack(app apps.AppCode) AppInstance {
	name := random.Name(random.WithPrefix(string(app)))
	session := cf.Start("push", "--no-start", "-p", app.Dir(), name)
	return waitForAppPush(session, name)
}

func appBuildAndPushUnstartedBinaryBuildpack(app apps.AppCode) AppInstance {
	name := random.Name(random.WithPrefix(string(app)))
	appDir := appBuild(app.Dir())
	defer os.RemoveAll(appDir)

	session := cf.Start("push", "--no-start", "-b", "binary_buildpack", "-m", "50MB", "-p", appDir, name)
	return waitForAppPush(session, name)
}

func appBuild(source string) string {
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

func waitForAppPush(session *Session, name string) AppInstance {
	Eventually(session, cfOperationWaitTime).Should(Exit())

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to push app. Getting logs...")
		cf.Run("logs", name, "--recent")
		Fail("App failed to push")
	}

	return AppInstance{name: name}
}

func waitForAppDelete(session *Session, name string) AppInstance {
	Eventually(session, cfOperationWaitTime).Should(Exit())

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to delete app. Getting logs...")
		cf.Run("logs", name, "--recent")
		Fail("App failed to delete")
	}

	return AppInstance{name: name}
}
