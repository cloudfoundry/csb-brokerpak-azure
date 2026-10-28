package integration_test

import (
	"fmt"

	testframework "github.com/cloudfoundry/cloud-service-broker/v2/brokerpaktestframework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AKS default expressions", Label("AKS"), func() {
	BeforeEach(func() {
		Expect(mockTerraform.SetTFState([]testframework.TFStateValue{})).To(Succeed())
	})

	AfterEach(func() {
		Expect(mockTerraform.Reset()).To(Succeed())
	})

	It("evaluates generated names through the real broker before Terraform", func() {
		id, err := broker.Provision("csb-azure-aks", "sandbox-3-node", map[string]any{})
		Expect(err).NotTo(HaveOccurred())
		expected := fmt.Sprintf("aks-%s", id)
		Expect(expected).To(HaveLen(40))
		Expect(mockTerraform.FirstTerraformInvocationVars()).To(SatisfyAll(
			HaveKeyWithValue("instance_name", expected),
			HaveKeyWithValue("resource_group", expected),
			HaveKeyWithValue("private_cluster_enabled", true),
		))
	})
})
