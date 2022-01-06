package servicekeys

import "acceptancetests/helpers/cf"

func (s *ServiceKey) Delete() {
	cf.Run("delete-service-key", "-f", s.serviceInstanceName, s.name)
}
