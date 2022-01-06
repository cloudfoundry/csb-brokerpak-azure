package services

import "acceptancetests/helpers/servicekeys"

func (s *ServiceInstance) CreateServiceKey() *servicekeys.ServiceKey {
	return servicekeys.Create(s.Name)
}
