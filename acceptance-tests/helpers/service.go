package helpers

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
	"encoding/json"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type ServiceInstance struct {
	offering string
	name     string
}

func CreateServiceFromBroker(offering, plan, broker string, parameters ...interface{}) ServiceInstance {
	name := random.Name(random.WithPrefix(offering, plan))
	createCommandTimeout := 5 * time.Minute // MASB is slow to start creation
	args := []string{"create-service", offering, plan, name, "-b", broker}
	if cf.Version() == cf.VersionV8 {
		args = append(args, "--wait")
		createCommandTimeout = time.Hour
	}
	args = append(args, serviceParameters(parameters)...)

	session := cf.Start(args...)
	Eventually(session, createCommandTimeout).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", name)
		Expect(out).NotTo(MatchRegexp(`status:\s+create failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+create succeeded`))

	return ServiceInstance{
		name:     name,
		offering: offering,
	}
}

func CreateService(offering, plan string, parameters ...interface{}) ServiceInstance {
	switch cf.Version() {
	case cf.VersionV8:
		return createServiceWithWait(offering, plan, parameters...)
	default:
		return createServiceWithPoll(offering, plan, parameters...)
	}
}

func createServiceWithWait(offering, plan string, parameters ...interface{}) ServiceInstance {
	name := random.Name(random.WithPrefix(offering, plan))
	args := append([]string{"create-service", offering, plan, name, "--wait"}, serviceParameters(parameters)...)

	session := cf.Start(args...)
	Eventually(session, time.Hour).Should(Exit(0), func() string {
		out, _ := cf.Run("service", name)
		return out
	})

	return ServiceInstance{
		name:     name,
		offering: offering,
	}
}

func createServiceWithPoll(offering, plan string, parameters ...interface{}) ServiceInstance {
	name := random.Name(random.WithPrefix(offering, plan))
	args := append([]string{"create-service", offering, plan, name}, serviceParameters(parameters)...)

	session := cf.Start(args...)
	Eventually(session, 5*time.Minute).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", name)
		Expect(out).NotTo(MatchRegexp(`status:\s+create failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+create succeeded`))

	return ServiceInstance{
		name:     name,
		offering: offering,
	}
}

func (s ServiceInstance) UpdateService(parameters ...string) {
	switch cf.Version() {
	case cf.VersionV8:
		s.updateServiceWithWait(parameters...)
	default:
		s.updateServiceWithPoll(parameters...)
	}
}

func (s ServiceInstance) updateServiceWithWait(parameters ...string) {
	args := append([]string{"update-service", s.name, "--wait"}, parameters...)

	session := cf.Start(args...)
	Eventually(session, time.Hour).Should(Exit(0), func() string {
		out, _ := cf.Run("service", s.name)
		return out
	})
}

func (s ServiceInstance) updateServiceWithPoll(parameters ...string) {
	args := append([]string{"update-service", s.name}, parameters...)

	session := cf.Start(args...)
	Eventually(session, 5*time.Minute).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("service", s.name)
		Expect(out).NotTo(MatchRegexp(`status:\s+update failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+update succeeded`))
}

func (s ServiceInstance) Delete() {
	args := []string{"delete-service", "-f", s.name}
	deleteCommandTimeout := time.Minute
	if cf.Version() == cf.VersionV8 {
		args = append(args, "--wait")
		deleteCommandTimeout = time.Hour
	}

	session := cf.Start(args...)
	Eventually(session, deleteCommandTimeout).Should(Exit(0))

	Eventually(func() string {
		out, _ := cf.Run("services")
		return out
	}, time.Hour, 30*time.Second).ShouldNot(ContainSubstring(s.name))
}

func (s ServiceInstance) Bind(app *apps.App, parameters ...interface{}) Binding {
	name := random.Name()
	args := []string{"bind-service", app.Name, s.name, "--binding-name", name}
	args = append(args, serviceParameters(parameters)...)
	cf.Run(args...)

	return Binding{
		serviceInstance: s,
		bindingName:     name,
		appInstance:     app,
	}
}

func (s ServiceInstance) Unbind(app *apps.App) {
	args := []string{"unbind-service", app.Name, s.name}
	cf.Run(args...)
}

func (s ServiceInstance) CreateKey() ServiceKey {
	name := random.Name()
	cf.Run("create-service-key", s.name, name)

	return ServiceKey{
		name:            name,
		serviceInstance: s,
	}
}

func (s ServiceInstance) Name() string {
	return s.name
}

func serviceParameters(parameters []interface{}) []string {
	if len(parameters) > 0 {
		switch p := parameters[0].(type) {
		case string:
			return []string{"-c", p}
		default:
			params, err := json.Marshal(p)
			Expect(err).NotTo(HaveOccurred())
			return []string{"-c", string(params)}
		}
	}

	return []string{}
}
