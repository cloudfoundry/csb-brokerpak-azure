package main

import (
	"flag"

	"csbbrokerpakazure/acceptance-tests/helpers/brokerpaks"
)

func main() {
	var version, dir string
	flag.StringVar(&version, "version", "", "version to upgrade from")
	flag.StringVar(&dir, "dir", "", "directory to install to")
	flag.Parse()

	if version == "" {
		version = brokerpaks.LatestVersion()
	}

	if dir == "" {
		dir = brokerpaks.TargetDir(version)
	}

	brokerpaks.DownloadBrokerpak(version, dir)
}
