//go:build tools
// +build tools

package tools

import (
	_ "github.com/cloudfoundry/cloud-service-broker/v2"
)

// This file imports the Cloud Service Broker in order to pin the version
// For other tools, use "go get -tool"
