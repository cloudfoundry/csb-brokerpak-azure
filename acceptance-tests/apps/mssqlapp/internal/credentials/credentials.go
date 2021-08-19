package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

func Read() (string, error) {
	app, err := cfenv.Current()
	if err != nil {
		return "", fmt.Errorf("error reading app env: %w", err)
	}
	if svs, err := app.Services.WithTag("mssql"); err == nil {
		return readService(svs)
	}
	if svs, err := app.Services.WithLabel("azure-sqldb"); err == nil {
		return readService(svs)
	}

	return "", fmt.Errorf("error reading MSSQL service details")
}

func readService(svs []cfenv.Service) (string, error) {
	var c Config
	if err := mapstructure.Decode(svs[0].Credentials, &c); err != nil {
		return "", fmt.Errorf("failed to decode credentials: %w", err)
	}

	if !c.Valid() {
		return "", fmt.Errorf("parsed credentials are not valid")
	}

	return c.String(), nil
}
