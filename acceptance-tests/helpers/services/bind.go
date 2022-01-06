package services

import (
	"acceptancetests/helpers/apps"
	"acceptancetests/helpers/bindings"
)

func (s *ServiceInstance) Bind(app *apps.App) *bindings.Binding {
	return bindings.Bind(s.Name, app.Name)
}
