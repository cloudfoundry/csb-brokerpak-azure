package aksvalidation

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Execute the production variable declarations without a provider or cloud API.
// This tests OpenTofu's actual evaluation, rather than duplicating CIDR rules.
func TestManagementCIDRs(t *testing.T) {
	variables, err := os.ReadFile("../../terraform/azure-aks/provision/variables.tf")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		private bool
		cidrs   []string
		valid   bool
	}{
		{"private_empty", true, []string{}, true},
		{"public_host", false, []string{"192.0.2.10/32"}, true},
		{"public_multiple_hosts", false, []string{"192.0.2.10/32", "198.51.100.7/32"}, true},
		{"private_with_host", true, []string{"192.0.2.10/32"}, false},
		{"public_empty", false, []string{}, false},
		{"world", false, []string{"0.0.0.0/0"}, false},
		{"subnet", false, []string{"192.0.2.0/24"}, false},
		{"ipv6", false, []string{"2001:db8::1/128"}, false},
		{"malformed", false, []string{"not-a-cidr/32"}, false},
		{"invalid_octet", false, []string{"999.0.2.10/32"}, false},
		{"missing_prefix", false, []string{"192.0.2.10"}, false},
		{"noncanonical", false, []string{"192.000.2.10/32"}, false},
		{"mixed_invalid", false, []string{"192.0.2.10/32", "0.0.0.0/0"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testManagementCIDRs(t, variables, tc.private, tc.cidrs, tc.valid)
		})
	}
}

func testManagementCIDRs(t *testing.T, variables []byte, private bool, cidrs []string, valid bool) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "variables.tf"), variables, 0600); err != nil {
		t.Fatal(err)
	}
	inputs := map[string]any{
		"instance_name": "csb-aks-test", "location": "eastus2", "resource_group": "csb-aks-test",
		"ttl_hours": 8, "node_count": 3, "node_vm_size": "Standard_D2_v5",
		"private_cluster_enabled": private, "authorized_api_server_cidrs": cidrs,
		"vnet_address_space": []string{"10.240.0.0/16"}, "subnet_address_prefixes": []string{"10.240.0.0/20"},
		"service_cidr": "10.241.0.0/16", "dns_service_ip": "10.241.0.10", "pod_cidr": "10.244.0.0/16",
		"labels": map[string]string{}, "cf_context_json": "{}", "cf_originating_identity_json": "{}",
		"azure_tenant_id": "test", "azure_subscription_id": "test", "azure_client_id": "test", "azure_client_secret": "test",
	}
	data, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "test.auto.tfvars.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("terraform", "plan", "-input=false", "-lock=false", "-no-color")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if (err == nil) != valid {
		t.Fatalf("expected valid=%v; error=%v\n%s", valid, err, output)
	}
	if !valid && !bytes.Contains(output, []byte("Invalid value for variable")) {
		t.Fatalf("expected input-validation failure, not an unrelated command error:\n%s", output)
	}
}
