package brokers

import (
	"csbbrokerpakazure/acceptance-tests/helpers/cf"
)

func (b *Broker) Delete() {
	// This is implicit when deleting the app, but sometimes that fails, so this ensures the resource is freed
	cf.Run("unbind-service", b.Name, "csb-sql")

	cf.Run("delete-service-broker", b.Name, "-f")
	b.app.Delete()
}
