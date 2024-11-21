// Package environment manages environment variables
package environment

import (
	"os"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/gomega"
)

type Metadata struct {
	ResourceGroup string `jsonry:"name"`
	PublicIP      string `jsonry:"v2.vm.ssh_ip"`
}

func ReadMetadata() Metadata {
	file := os.Getenv("ENVIRONMENT_LOCK_METADATA")
	Expect(file).NotTo(BeEmpty(), "You must set the ENVIRONMENT_LOCK_METADATA environment variable")

	contents, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())

	var metadata Metadata
	Expect(jsonry.Unmarshal(contents, &metadata)).NotTo(HaveOccurred())
	Expect(metadata.ResourceGroup).NotTo(BeEmpty())
	return metadata
}
