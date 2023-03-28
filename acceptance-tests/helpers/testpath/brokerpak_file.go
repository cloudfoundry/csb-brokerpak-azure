package testpath

import (
	"fmt"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func BrokerpakFile(parts ...string) string {
	GinkgoHelper()

	r := BrokerpakRoot()
	p := filepath.Join(append([]string{r}, parts...)...)
	Expect(p).To(BeAnExistingFile(), func() string {
		return fmt.Sprintf("could not find file %q in brokerpak %q", path.Join(parts...), r)
	})

	return p
}
