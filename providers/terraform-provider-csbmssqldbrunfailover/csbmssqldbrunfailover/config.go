package csbmssqldbrunfailover

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// getIdentifier gets a string configuration value and validates that it's a valid identifier
func getIdentifier(d *schema.ResourceData, key string) (string, diag.Diagnostics) {
	s := d.Get(key).(string)
	if s == "" {
		return "", diag.Errorf("empty value for identifier %q", key)
	}

	return s, nil
}
