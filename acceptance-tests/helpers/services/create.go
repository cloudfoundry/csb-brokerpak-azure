package services

import (
	"acceptancetests/helpers/brokers"
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
	"encoding/json"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type ServiceInstance struct {
	Name string
}

type config struct {
	name              string
	serviceBrokerName string
	parameters        string
}

type Option func(*config)

func CreateInstance(offering, plan string, opts ...Option) *ServiceInstance {
	cfg := defaultConfig(offering, plan, opts...)
	args := []string{
		"create-service",
		offering,
		plan,
		cfg.name,
		"-b",
		cfg.serviceBrokerName,
	}

	if cfg.parameters != "" {
		args = append(args, "-c", cfg.parameters)
	}

	switch cf.Version() {
	case cf.VersionV8:
		createInstanceWithWait(cfg.name, args)
	default:
		createInstanceWithPoll(cfg.name, args)
	}

	return &ServiceInstance{Name: cfg.name}
}

func createInstanceWithWait(name string, args []string) {
	args = append(args, "--wait")
	session := cf.Start(args...)
	Eventually(session, time.Hour).Should(Exit(0), func() string {
		out, _ := cf.Run("service", name)
		return out
	})
}

func createInstanceWithPoll(name string, args []string) {
	session := cf.Start(args...)
	Eventually(session, 5*time.Minute).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", name)
		Expect(out).NotTo(MatchRegexp(`status:\s+create failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+create succeeded`))
}

func WithDefaultBroker() Option {
	return func(c *config) {
		c.serviceBrokerName = brokers.DefaultBrokerName()
	}
}

func WithBroker(broker *brokers.Broker) Option {
	return func(c *config) {
		c.serviceBrokerName = broker.Name
	}
}

func WithParameters(parameters interface{}) Option {
	return func(c *config) {
		switch p := parameters.(type) {
		case string:
			c.parameters = p
		default:
			params, err := json.Marshal(p)
			Expect(err).NotTo(HaveOccurred())
			c.parameters = string(params)
		}
	}
}

func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

func WithOptions(opts ...Option) Option {
	return func(c *config) {
		for _, o := range opts {
			o(c)
		}
	}
}

func defaultConfig(offering, plan string, opts ...Option) config {
	var cfg config
	WithOptions(append([]Option{
		WithDefaultBroker(),
		WithName(random.Name(random.WithPrefix(offering, plan))),
	}, opts...)...)(&cfg)
	return cfg
}
