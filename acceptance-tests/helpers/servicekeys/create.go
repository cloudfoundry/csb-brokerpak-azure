package servicekeys

import (
	"acceptancetests/helpers/cf"
	"acceptancetests/helpers/random"
)

type ServiceKey struct {
	name                string
	serviceInstanceName string
}

func Create(serviceInstanceName string) *ServiceKey {
	name := random.Name()
	cf.Run("create-service-key", serviceInstanceName, name)

	return &ServiceKey{
		name:                name,
		serviceInstanceName: serviceInstanceName,
	}
}
