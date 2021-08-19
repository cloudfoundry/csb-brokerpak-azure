package helpers

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

func SetBrokerEnv(key string, value interface{}) {
	const broker = "cloud-service-broker"
	var val string
	switch v := value.(type) {
	case string:
		val = v
	default:
		data, err := json.Marshal(v)
		Expect(err).NotTo(HaveOccurred())
		val = string(data)
	}

	CF("set-env", broker, key, val)
	CF("restart", broker)
}
