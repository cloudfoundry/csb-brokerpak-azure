package credentials

import (
	"fmt"
	"reflect"
	"strings"
)

type Config struct {
	UserID   string `mapstructure:"username" config:"user id"`
	Password string `mapstructure:"password" config:"password"`
	Server   string `mapstructure:"hostname" config:"server"`
	Port     int    `mapstructure:"port" config:"port"`
	Database string `mapstructure:"name" config:"database"`
}

func (c Config) Valid() bool {
	for _, v := range c.toMap() {
		if reflect.ValueOf(v).IsZero() {
			return false
		}
	}
	return true
}

func (c Config) String() string {
	params := c.toMap()
	params["encrypt"] = true

	var s strings.Builder
	for k, v := range params {
		s.WriteString(k)
		switch t := v.(type) {
		case int:
			s.WriteString(fmt.Sprintf("=%d; ", t))
		case bool:
			s.WriteString(fmt.Sprintf("=%t; ", t))
		default:
			s.WriteString(fmt.Sprintf("=%s; ", t))
		}
	}
	return s.String()
}

func (c Config) toMap() map[string]any {
	m := make(map[string]any)
	v := reflect.ValueOf(c)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Tag.Get("config")
		value := v.Field(i).Interface()
		m[key] = value
	}
	return m
}
