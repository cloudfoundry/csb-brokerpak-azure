package helpers

import (
	"time"

	. "github.com/onsi/gomega"
)

func DeleteService(name string) {
	CF("delete-service", "-f", name)
	Eventually(func() string {
		out, _ := CF("services")
		return out
	}, 30*time.Minute, 30*time.Second).ShouldNot(ContainSubstring(name))
}
