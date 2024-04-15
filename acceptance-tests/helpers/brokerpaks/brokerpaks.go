// Package brokerpaks is used for downloading brokerpaks and associated resources for upgrade tests
package brokerpaks

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const brokerpak = "cloudfoundry/csb-brokerpak-azure"

// DownloadBrokerpak will download the brokerpak of the specified
// version and return the directory where it has been placed.
// The download is skipped if it has previously been downloaded.
// Includes downloading the corresponding broker and ".envrc" file
func DownloadBrokerpak(version, dir string) string {
	// Brokerpak
	basename := fmt.Sprintf("azure-services-%s.brokerpak", version)
	uri := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", brokerpak, version, basename)
	downloadUnlessCached(dir, basename, uri)

	// ".envrc" file
	envrcURI := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/.envrc", brokerpak, version)
	downloadUnlessCached(dir, ".envrc", envrcURI)

	// broker
	brokerVersion := readBrokerVersion(version)
	brokerURI := fmt.Sprintf("https://github.com/cloudfoundry/cloud-service-broker/releases/download/%s/cloud-service-broker.linux", brokerVersion)
	downloadUnlessCached(dir, "cloud-service-broker", brokerURI)
	if err := os.Chmod(filepath.Join(dir, "cloud-service-broker"), 0777); err != nil {
		panic(err)
	}

	return dir
}

// readBrokerVersion will use the specified brokerpak version to determine the corresponding broker version
func readBrokerVersion(version string) string {
	body := newClient().get(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/go.mod", brokerpak, version), "text/plain")
	defer body.Close()
	data := must(io.ReadAll(body))

	matches := regexp.MustCompile(`(?m)^\s*github\.com/cloudfoundry/cloud-service-broker(/v\d+)?\s+(\S+)\s*$`).FindSubmatch(data)
	if len(matches) != 3 {
		panic(fmt.Sprintf("Could not extract CSB version from go.mod file: %q", data))
	}

	brokerVersion := string(matches[2])
	fmt.Printf("Brokerpak version %q uses broker version %q\n", version, brokerVersion)
	return brokerVersion
}

// downloadUnlessCached will download a file to a known location, unless it's already there
func downloadUnlessCached(dir, basename, uri string) {
	target := filepath.Join(dir, basename)

	_, err := os.Stat(target)
	switch err {
	case nil:
		fmt.Printf("Found %q cached at %q.\n", uri, target)
	default:
		fmt.Printf("Downloading %q to %q.\n", uri, target)
		newClient().download(target, uri)
	}
}

// TargetDir will determine the target directory for a version and make sure that it exists
func TargetDir(version string) string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(pwd, "versions", version)
	if err := os.MkdirAll(dir, 0777); err != nil {
		panic(err)
	}

	return dir
}
