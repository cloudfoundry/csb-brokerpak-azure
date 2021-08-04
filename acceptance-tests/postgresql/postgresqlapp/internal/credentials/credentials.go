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
	svs, err := app.Services.WithTag("postgresql")
	if err != nil {
		return "", fmt.Errorf("error reading PostgreSQL service details")
	}

	var m struct {
		URI string `mapstructure:"uri"`
	}

	if err := mapstructure.Decode(svs[0].Credentials, &m); err != nil {
		return "", fmt.Errorf("failed to decode credentials: %w", err)
	}

	if m.URI == "" {
		return "", fmt.Errorf("parsed credentials are not valid")
	}

	return m.URI, nil
}
