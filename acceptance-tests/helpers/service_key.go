package helpers

import (
	"acceptancetests/helpers/cf"
	"encoding/json"
	"reflect"
	"strings"

	. "github.com/onsi/gomega"
)

type ServiceKey struct {
	name            string
	serviceInstance ServiceInstance
}

func (k ServiceKey) Get(receiver interface{}) {
	Expect(reflect.ValueOf(receiver).Kind()).To(Equal(reflect.Ptr), "receiver must be a pointer")
	out, _ := cf.Run("service-key", k.serviceInstance.name, k.name)
	start := strings.Index(out, "{")
	Expect(start).To(BeNumerically(">", 0), "could not find start of JSON")
	data := []byte(out[start:])

	if cf.Version() == cf.VersionV8 {
		var wrapper struct {
			Credentials interface{} `json:"credentials"`
		}
		err := json.Unmarshal(data, &wrapper)
		Expect(err).NotTo(HaveOccurred())
		data, err = json.Marshal(wrapper.Credentials)
		Expect(err).NotTo(HaveOccurred())
	}

	err := json.Unmarshal(data, receiver)
	Expect(err).NotTo(HaveOccurred())
}

func (k ServiceKey) Delete() {
	cf.Run("delete-service-key", "-f", k.serviceInstance.name, k.name)
}
