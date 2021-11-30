package apps

import (
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const pushWaitTime = 20 * time.Minute

type config struct {
	name      string
	start     bool
	buildpack string
	memory    string
	manifest  string
	variables map[string]string
	dir       dir
}

type Option func(*config)

func Push(opts ...Option) App {
	var c config
	defaults := []Option{WithName(random.Name(random.WithPrefix("app")))}
	WithOptions(append(defaults, opts...)...)(&c)

	cmd := []string{"push"}
	if !c.start {
		cmd = append(cmd, "--no-start")
	}
	if c.buildpack != "" {
		cmd = append(cmd, "-b", c.buildpack)
	}
	if c.memory != "" {
		cmd = append(cmd, "-m", c.memory)
	}
	if c.manifest != "" {
		cmd = append(cmd, "-f", c.manifest)
	}

	for k, v := range c.variables {
		cmd = append(cmd, "--var", fmt.Sprintf("%s=%s", k, v))
	}

	if c.dir.path() == "" {
		Fail("App directory must be specified")
	}
	cmd = append(cmd, "-p", c.dir.path())
	defer c.dir.cleanup()

	if c.name == "" {
		Fail("App name must be specified")
	}
	cmd = append(cmd, c.name)

	session := cf.Start(cmd...)
	Eventually(session, pushWaitTime).Should(gexec.Exit())
	checkSuccess(session.ExitCode(), c.name)

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to push app. Getting logs...")
		cf.Run("logs", c.name, "--recent")
		Fail("App failed to push")
	}

	return App{
		Name: c.name,
		URL:  url(c.name),
	}
}

func WithBinaryBuildpack() Option {
	return func(c *config) {
		c.buildpack = "binary_buildpack"
		c.memory = "50MB"
	}
}

func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

func WithDir(dir string) Option {
	return func(c *config) {
		c.dir = staticDir(dir)
	}
}

func WithManifest(manifest string) Option {
	return func(c *config) {
		c.manifest = manifest
	}
}

func WithVariable(key, value string) Option {
	return func(c *config) {
		if c.variables == nil {
			c.variables = make(map[string]string)
		}
		c.variables[key] = value
	}
}

func WithStartedState() Option {
	return func(c *config) {
		c.start = true
	}
}

func WithOptions(opts ...Option) Option {
	return func(c *config) {
		for _, o := range opts {
			o(c)
		}
	}
}
