package cf

import (
	"strings"
)

type versionType int

const (
	VersionUnknown versionType = iota
	VersionLegacy
	VersionV8
)

var cachedVersion versionType

func Version() versionType {
	if cachedVersion == VersionUnknown {
		out, _ := Run("version")
		switch {
		case strings.HasPrefix(out, "cf version 8"):
			cachedVersion = VersionV8
		default:
			cachedVersion = VersionLegacy
		}
	}
	return cachedVersion
}
