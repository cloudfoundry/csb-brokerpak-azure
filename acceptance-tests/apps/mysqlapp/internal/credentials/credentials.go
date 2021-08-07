package credentials

import (
	"fmt"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
)

func Read() (*mysql.Config, error) {
	app, err := cfenv.Current()
	if err != nil {
		return nil, fmt.Errorf("error reading app env: %w", err)
	}
	svs, err := app.Services.WithTag("mysql")
	if err != nil {
		return nil, fmt.Errorf("error reading MySQL service details")
	}

	var m struct {
		Host     string `mapstructure:"hostname"`
		Database string `mapstructure:"name"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Port     int    `mapstructure:"port"`
	}

	if err := mapstructure.Decode(svs[0].Credentials, &m); err != nil {
		return nil, fmt.Errorf("failed to decode credentials: %w", err)
	}

	if m.Host == "" || m.Username == "" || m.Password == "" || m.Database == "" || m.Port == 0 {
		return nil, fmt.Errorf("parsed credentials are not valid")
	}

	c := mysql.NewConfig()
	c.TLSConfig = "true"
	c.Net = "tcp"
	c.Addr = m.Host
	c.User = m.Username
	c.Passwd = m.Password
	c.DBName = m.Database

	return c, nil
}
