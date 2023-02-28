package apps

import (
	"fmt"

	"csbbrokerpakazure/acceptance-tests/helpers/cf"

	. "github.com/onsi/ginkgo/v2"
)

func checkSuccess(code int, name string) {
	if code != 0 {
		fmt.Fprintln(GinkgoWriter, "Operation FAILED. Getting logs...")
		cf.Run("logs", name, "--recent")
		Fail("App operation failed")
	}
}
