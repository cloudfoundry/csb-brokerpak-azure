package helpers

import (
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
	out, _ := CF("service-key", k.serviceInstance.name, k.name)
	start := strings.Index(out, "{")
	Expect(start).To(BeNumerically(">", 0), "could not find start of JSON")
	err := json.Unmarshal([]byte(out[start:]), receiver)
	Expect(err).NotTo(HaveOccurred())
}

func (k ServiceKey) Delete() {
	CF("delete-service-key", "-f", k.serviceInstance.name, k.name)
}
