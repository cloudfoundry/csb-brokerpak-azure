package credentials

import "fmt"

type Config struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Hostname string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"name"`
}

func (c Config) Valid() bool {
	return c.Username != "" && c.Password != "" && c.Hostname != "" && c.Port != 0 && c.Database != ""
}

func (c Config) String() string {
	return fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;", c.Hostname, c.Username, c.Password, c.Port, c.Database)
}
