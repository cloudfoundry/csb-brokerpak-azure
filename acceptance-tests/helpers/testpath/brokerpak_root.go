package testpath

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
)

// BrokerpakRoot searches upwards from the current working directory to find the root path of the brokerpak
// Fails the test if not found
func BrokerpakRoot() string {
	ginkgo.GinkgoHelper()

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	d, err := filepath.Abs(cwd)
	if err != nil {
		panic(err)
	}

	for {
		switch {
		case Exists(filepath.Join(d, "manifest.yml")) && Exists(filepath.Join(d, "acceptance-tests")):
			return d
		case d == "/":
			ginkgo.Fail(fmt.Sprintf("could not determine brokerpak root from %q", cwd))
		default:
			d = filepath.Dir(d)
		}
	}
}
