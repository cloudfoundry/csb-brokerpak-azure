// Package testpath provides path utilities for tests
package testpath

import (
	"os"
)

// Exists returns whether a path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
