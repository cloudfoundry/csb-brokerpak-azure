package helpers

func Bind(app, serviceInstance string) string {
	name := RandomName()
	CF("bind-service", app, serviceInstance, "--binding-name", name)
	return name
}
