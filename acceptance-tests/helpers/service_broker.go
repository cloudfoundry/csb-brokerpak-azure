package helpers

import (
	"fmt"
	"os"
)

func DefaultBrokerName() string {
	if v, ok := os.LookupEnv("BROKER_NAME"); ok {
		return v
	} else if v, ok := os.LookupEnv("USER"); ok {
		return fmt.Sprintf("csb-%s", v)
	} else {
		panic("could not compute default broker name")
	}
}
