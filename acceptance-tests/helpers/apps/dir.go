package apps

import (
	"os"

	. "github.com/onsi/gomega"
)

type dir interface {
	path() string
	cleanup()
}

type staticDir string

func (s staticDir) path() string {
	return string(s)
}

func (staticDir) cleanup() {}

type tmpDir string

func newTmpDir() tmpDir {
	dir, err := os.MkdirTemp("", "")
	Expect(err).NotTo(HaveOccurred())
	return tmpDir(dir)
}

func (t tmpDir) path() string {
	return string(t)
}

func (t tmpDir) cleanup() {
	Expect(os.RemoveAll(t.path())).NotTo(HaveOccurred())
}
