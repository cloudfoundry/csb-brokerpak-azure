package brokers

import (
	"os"
	"path/filepath"

	"csbbrokerpakazure/acceptance-tests/helpers/apps"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

func newManifest(opts ...manifestOption) string {

	m := manifestModel{
		Version: 1,
		Applications: []applicationModel{
			{
				Command:     "./cloud-service-broker serve",
				Memory:      "750MB",
				Disk:        "2G",
				Buildpacks:  []string{"binary_buildpack"},
				RandomRoute: true,
				Environment: make(map[string]string),
			},
		},
	}

	for _, o := range opts {
		o(&m)
	}

	data, err := yaml.Marshal(m)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	dir := ginkgo.GinkgoT().TempDir()
	path := filepath.Join(dir, "manifest.yml")
	gomega.Expect(os.WriteFile(path, data, 0666)).To(gomega.Succeed())

	return path
}

type manifestOption func(*manifestModel)

type manifestModel struct {
	Version      int                `yaml:"version"`
	Applications []applicationModel `yaml:"applications"`
}

type applicationModel struct {
	Name        string            `yaml:"name"`
	Command     string            `yaml:"command"`
	Memory      string            `yaml:"memory,omitempty"`
	Disk        string            `yaml:"disk_quota,omitempty"`
	Buildpacks  []string          `yaml:"buildpacks,omitempty"`
	RandomRoute bool              `yaml:"random-route"`
	Environment map[string]string `yaml:"env,omitempty"`
}

func withName(name string) manifestOption {
	return func(m *manifestModel) {
		m.Applications[0].Name = name
	}
}
func withCustomStartCommand(command string) manifestOption {
	return func(m *manifestModel) {
		m.Applications[0].Command = command
	}
}

func withEnv(env ...apps.EnvVar) manifestOption {
	return func(m *manifestModel) {
		for _, e := range env {
			m.Applications[0].Environment[e.Name] = e.ValueString()
		}
	}
}
