package services

import (
	"csbbrokerpakazure/acceptance-tests/helpers/apps"
	"csbbrokerpakazure/acceptance-tests/helpers/bindings"
)

func (s *ServiceInstance) Bind(app *apps.App) *bindings.Binding {
	return bindings.Bind(s.Name, app.Name)
}
