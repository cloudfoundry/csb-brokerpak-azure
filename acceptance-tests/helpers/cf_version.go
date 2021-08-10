package helpers

import (
	"strings"
)

type cfVersionType int

const (
	cfVersionUnknown cfVersionType = iota
	cfVersionLegacy
	cfVersionV8
)

var cachedCFVersion cfVersionType

func cfVersion() cfVersionType {
	if cachedCFVersion == cfVersionUnknown {
		out, _ := CF("version")
		switch {
		case strings.HasPrefix(out, "cf version 8"):
			cachedCFVersion = cfVersionV8
		default:
			cachedCFVersion = cfVersionLegacy
		}
	}
	return cachedCFVersion
}
