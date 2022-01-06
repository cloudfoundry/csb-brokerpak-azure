package servicekeys

import (
	"acceptancetests/helpers/cf"
	"encoding/json"
	"reflect"
	"strings"

	. "github.com/onsi/gomega"
)

func (s *ServiceKey) Get(receiver interface{}) {
	Expect(reflect.ValueOf(receiver).Kind()).To(Equal(reflect.Ptr), "receiver must be a pointer")
	out, _ := cf.Run("service-key", s.serviceInstanceName, s.name)
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

	Expect(json.Unmarshal(data, receiver)).NotTo(HaveOccurred())
}
