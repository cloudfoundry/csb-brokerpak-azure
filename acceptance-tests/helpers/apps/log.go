package apps

import (
	"acceptancetests/helpers/cf"
	"fmt"

	. "github.com/onsi/ginkgo"
)

func checkSuccess(code int, name string) {
	if code != 0 {
		fmt.Fprintln(GinkgoWriter, "Operation FAILED. Getting logs...")
		cf.Run("logs", name, "--recent")
		Fail("App operation failed")
	}
}
