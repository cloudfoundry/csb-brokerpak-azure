package credentials

import (
	"fmt"
	"log"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/mitchellh/mapstructure"
)

func Read() (string, error) {
	app, err := cfenv.Current()
	if err != nil {
		return "", fmt.Errorf("error reading app env: %w", err)
	}
	if svs, err := app.Services.WithTag("mssql"); err == nil {
		log.Println("found tag: mssql")
		return readService(svs)
	}
	if svs, err := app.Services.WithLabel("azure-sqldb"); err == nil {
		log.Println("found label: azure-sqldb")
		return readService(svs)
	}
	if svs, err := app.Services.WithLabel("azure-sqldb-failover-group"); err == nil {
		log.Println("found label: azure-sqldb-failover-group")
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
