package bindings

import (
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
)

type Binding struct {
	name                string
	serviceInstanceName string
	appName             string
}

func Bind(serviceInstanceName, appName string) *Binding {
	name := random.Name()
	cf.Run("bind-service", appName, serviceInstanceName, "--binding-name", name)
	return &Binding{
		name:                name,
		serviceInstanceName: serviceInstanceName,
		appName:             appName,
	}
}
