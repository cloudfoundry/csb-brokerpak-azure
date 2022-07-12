// Package apps manages the test app lifecycle
package apps

type App struct {
	Name      string
	URL       string
	start     bool
	buildpack string
	memory    string
	disk      string
	manifest  string
	variables map[string]string
	dir       dir
}
