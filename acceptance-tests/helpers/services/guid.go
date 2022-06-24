package services

import (
	"csbbrokerpakazure/acceptance-tests/helpers/cf"
	"strings"
)

func (s *ServiceInstance) GUID() string {
	if s.guid == "" {
		out, _ := cf.Run("service", s.Name, "--guid")
		s.guid = strings.TrimSpace(out)
	}

	return s.guid
}
