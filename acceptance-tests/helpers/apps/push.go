package apps

import (
	"csbbrokerpakazure/acceptance-tests/helpers/cf"
	"csbbrokerpakazure/acceptance-tests/helpers/random"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const pushWaitTime = 20 * time.Minute

type Option func(*App)

func Push(opts ...Option) *App {
	defaults := []Option{
		WithName(random.Name(random.WithPrefix("app"))),
		WithMemory("100MB"),
		WithDisk("250MB"),
	}
	app := App{}
	app.Push(append(defaults, opts...)...)
	return &app
}

func (a *App) Push(opts ...Option) {
	WithOptions(opts...)(a)

	cmd := []string{"push"}
	if !a.start {
		cmd = append(cmd, "--no-start")
	}
	if a.buildpack != "" {
		cmd = append(cmd, "-b", a.buildpack)
	}
	if a.memory != "" {
		cmd = append(cmd, "-m", a.memory)
	}
	if a.disk != "" {
		cmd = append(cmd, "-k", a.disk)
	}
	if a.manifest != "" {
		cmd = append(cmd, "-f", a.manifest)
	}

	for k, v := range a.variables {
		cmd = append(cmd, "--var", fmt.Sprintf("%s=%s", k, v))
	}

	if a.dir.path() == "" {
		Fail("App directory must be specified")
	}
	cmd = append(cmd, "-p", a.dir.path())
	defer a.dir.cleanup()

	if a.Name == "" {
		Fail("App name must be specified")
	}
	cmd = append(cmd, a.Name)

	session := cf.Start(cmd...)
	Eventually(session, pushWaitTime).Should(gexec.Exit())
	checkSuccess(session.ExitCode(), a.Name)

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to push app. Getting logs...")
		cf.Run("logs", a.Name, "--recent")
		Fail("App failed to push")
	}

	a.URL = url(a.Name)
}

func WithBinaryBuildpack() Option {
	return func(a *App) {
		a.buildpack = "binary_buildpack"
		a.memory = "50MB"
	}
}

func WithName(name string) Option {
	return func(a *App) {
		a.Name = name
	}
}

func WithDir(dir string) Option {
	return func(a *App) {
		a.dir = staticDir(dir)
	}
}

func WithManifest(manifest string) Option {
	return func(a *App) {
		a.manifest = manifest
	}
}

func WithVariable(key, value string) Option {
	return func(a *App) {
		if a.variables == nil {
			a.variables = make(map[string]string)
		}
		a.variables[key] = value
	}
}

func WithStartedState() Option {
	return func(a *App) {
		a.start = true
	}
}

func WithMemory(memory string) Option {
	return func(a *App) {
		a.memory = memory
	}
}

func WithDisk(disk string) Option {
	return func(a *App) {
		a.disk = disk
	}
}

func WithOptions(opts ...Option) Option {
	return func(a *App) {
		for _, o := range opts {
			o(a)
		}
	}
}
