package credentials

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/jsonry"
)

func Read() (string, error) {
	const variable = "VCAP_SERVICES"

	type MongoDBService struct {
		URI string `jsonry:"credentials.uri"`
	}

	var services struct {
		MongoDBServices []MongoDBService `jsonry:"csb-azure-mongodb"`
	}

	if err := jsonry.Unmarshal([]byte(os.Getenv(variable)), &services); err != nil {
		return "", fmt.Errorf("failed to parse %q: %w", variable, err)
	}

	switch len(services.MongoDBServices) {
	case 1: // ok
	case 0:
		return "", fmt.Errorf("unable to find `csb-azure-mongodb` in %q", variable)
	default:
		return "", fmt.Errorf("more than one entry for `csb-azure-mongodb` in %q", variable)
	}

	uri := services.MongoDBServices[0].URI
	if uri == "" {
		return "", fmt.Errorf("parsed credentials are not valid: %s", os.Getenv(variable))
	}

	return uri, nil
}
