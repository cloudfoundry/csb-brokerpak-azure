package credentials

import (
	"code.cloudfoundry.org/jsonry"
	"crypto/tls"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
)

func Read() (*redis.Options, error) {
	const variable = "VCAP_SERVICES"

	type RedisService struct {
		Host     string `jsonry:"credentials.host"`
		Password string `jsonry:"credentials.password"`
		TLSPort  int    `jsonry:"credentials.tls_port"`
	}

	var services struct {
		RedisServices []RedisService `jsonry:"csb-azure-redis"`
	}

	if err := jsonry.Unmarshal([]byte(os.Getenv(variable)), &services); err != nil {
		return nil, fmt.Errorf("failed to parse %q: %w", variable, err)
	}

	switch len(services.RedisServices) {
	case 1: // ok
	case 0:
		return nil, fmt.Errorf("unable to find `csb-azure-redis` in %q", variable)
	default:
		return nil, fmt.Errorf("more than one entry for `csb-azure-redis` in %q", variable)
	}

	r := services.RedisServices[0]
	if r.Host == "" || r.Password == "" || r.TLSPort == 0 {
		return nil, fmt.Errorf("parsed credentials are not valid: %s", os.Getenv(variable))
	}

	return &redis.Options{
		Addr:      fmt.Sprintf("%s:%d", services.RedisServices[0].Host, services.RedisServices[0].TLSPort),
		Password:  services.RedisServices[0].Password,
		DB:        0,
		TLSConfig: &tls.Config{},
	}, nil
}
