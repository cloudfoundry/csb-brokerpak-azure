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
	buildpack string
	memory    string
	dir       dir
}

type Option func(*config)

func Push(opts ...Option) App {
	var c config
	defaults := []Option{WithName(random.Name(random.WithPrefix("app")))}
	WithOptions(append(defaults, opts...)...)(&c)

	if c.dir.path() == "" {
		Fail("App directory must be specified")
	}

	cmd := []string{"push", "--no-start"}
	switch {
	case c.buildpack != "":
		cmd = append(cmd, "-b", c.buildpack)
	case c.memory != "":
		cmd = append(cmd, "-m", c.memory)
	}

	cmd = append(cmd, "-p", c.dir.path(), c.name)
	defer c.dir.cleanup()

	session := cf.Start(cmd...)
	Eventually(session, pushWaitTime).Should(gexec.Exit())
	checkSuccess(session.ExitCode(), c.name)

	if session.ExitCode() != 0 {
		fmt.Fprintf(GinkgoWriter, "FAILED to push app. Getting logs...")
		cf.Run("logs", c.name, "--recent")
		Fail("App failed to push")
	}

	return App{Name: c.name}
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

func WithOptions(opts ...Option) Option {
	return func(c *config) {
		for _, o := range opts {
			o(c)
		}
	}
}
