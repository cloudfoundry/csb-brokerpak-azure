package helpers

import (
	"acceptancetests/apps"
	"code.cloudfoundry.org/jsonry"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type AppInstance struct {
	name string
}

func PushBrokerApp(path string) {
	name := "cloud-service-broker"
	session := StartCF("push", "-p", path, "-f", "cf-manifest.yml", name, "--var", "app=cloud-service-broker")
	waitForAppPush(session, name)

	out, _ := CF("app", "--guid", name)
	guid := strings.TrimSpace(out)
	env, _ := CF("curl", fmt.Sprintf("/v3/apps/%s/env", guid))

	var receiver struct {
		UserName string `jsonry:"environment_variables.SECURITY_USER_NAME"`
		UserPassword string `jsonry:"environment_variables.SECURITY_USER_PASSWORD"`
		BrokerURL []string `jsonry:"application_env_json.VCAP_APPLICATION.application_uris[]"`
	}
	err := jsonry.Unmarshal([]byte(env), &receiver)
	Expect(err).NotTo(HaveOccurred())

	Expect(receiver.UserName).NotTo(BeEmpty())
	Expect(receiver.UserPassword).NotTo(BeEmpty())
	Expect(receiver.BrokerURL).To(HaveLen(1))


	session = StartCF("update-service-broker", "broker-cf-test", receiver.UserName, receiver.UserPassword, "https://" + receiver.BrokerURL[0])
	waitForBrokerUpdate(session, name)
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
	name := RandomName(string(app))
	session := StartCF("push", "--no-start", "-p", app.Dir(), name)
	return waitForAppPush(session, name)
}

func appBuildAndPushUnstartedBinaryBuildpack(app apps.AppCode) AppInstance {
	name := RandomName(string(app))
	appDir := appBuild(app.Dir())
	defer os.RemoveAll(appDir)

	session := StartCF("push", "--no-start", "-b", "binary_buildpack", "-m", "50MB", "-p", appDir, name)
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
	Eventually(session, 5*time.Minute).Should(Exit())

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to push app. Getting logs...")
		CF("logs", name, "--recent")
		Fail("App failed to push")
	}

	return AppInstance{name: name}
}

func waitForBrokerUpdate(session *Session, name string) AppInstance {
	Eventually(session, 5*time.Minute).Should(Exit())

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to update broker. Getting logs...")
		CF("logs", name, "--recent")
		Fail("Broker failed to update")
	}

	return AppInstance{name: name}
}
