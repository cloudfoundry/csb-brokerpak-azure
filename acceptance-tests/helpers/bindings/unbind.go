package bindings

import (
	"acceptancetests/helpers/cf"
)

func (b *Binding) Unbind() {
	cf.Run("unbind-service", b.appName, b.serviceInstanceName)
}
