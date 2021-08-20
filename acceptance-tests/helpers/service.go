package helpers

import (
	"encoding/json"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type ServiceInstance struct {
	offering string
	name     string
}

func CreateService(offering, plan string, parameters ...interface{}) ServiceInstance {
	name := RandomName(offering, plan)
	createCommandTimeout := 5 * time.Minute // MASB is slow to start creation
	args := []string{"create-service", offering, plan, name}
	if cfVersion() == cfVersionV8 {
		args = append(args, "--wait")
		createCommandTimeout = time.Hour
	}
	args = append(args, serviceParameters(parameters)...)

	session := StartCF(args...)
	Eventually(session, createCommandTimeout).Should(Exit(0))

	Eventually(func() string {
		out, _ := CF("service", name)
		Expect(out).NotTo(MatchRegexp(`status:\s+create failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+create succeeded`))

	return ServiceInstance{
		name:     name,
		offering: offering,
	}
}

func (s ServiceInstance) UpdateService(parameters ...string) {
	createCommandTimeout := time.Minute
	args := []string{"update-service", s.name}
	if cfVersion() == cfVersionV8 {
		args = append(args, "--wait")
		createCommandTimeout = time.Hour
	}
	args = append(args, parameters...)

	session := StartCF(args...)
	Eventually(session, createCommandTimeout).Should(Exit(0))

	Eventually(func() string {
		out, _ := CF("service", s.name)
		Expect(out).NotTo(MatchRegexp(`status:\s+update failed`))
		return out
	}, time.Hour, 30*time.Second).Should(MatchRegexp(`status:\s+update succeeded`))
}

func (s ServiceInstance) Delete() {
	args := []string{"delete-service", "-f", s.name}
	deleteCommandTimeout := time.Minute
	if cfVersion() == cfVersionV8 {
		args = append(args, "--wait")
		deleteCommandTimeout = time.Hour
	}

	session := StartCF(args...)
	Eventually(session, deleteCommandTimeout).Should(Exit(0))

	Eventually(func() string {
		out, _ := CF("services")
		return out
	}, 30*time.Minute, 30*time.Second).ShouldNot(ContainSubstring(s.name))
}

func (s ServiceInstance) Bind(app AppInstance, parameters ...interface{}) Binding {
	name := RandomName()
	args := []string{"bind-service", app.name, s.name, "--binding-name", name}
	args = append(args, serviceParameters(parameters)...)
	CF(args...)

	return Binding{
		serviceInstance: s,
		bindingName:     name,
		appInstance:     app,
	}
}

func (s ServiceInstance) CreateKey() ServiceKey {
	name := RandomName()
	CF("create-service-key", s.name, name)

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
