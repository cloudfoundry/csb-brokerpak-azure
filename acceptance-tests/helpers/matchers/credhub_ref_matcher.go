// Package matchers has custom Gomega matchers
package matchers

import "github.com/onsi/gomega"

var HaveCredHubRef = gomega.HaveKey("credhub-ref")
