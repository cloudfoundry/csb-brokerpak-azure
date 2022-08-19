package csbmssqldbrunfailover

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	maxLengthClientSecret = 1024
)

var (
	identifierRegexp = regexp.MustCompile(`^[\w_.-]{1,64}$`)
)

// getIdentifier gets a string configuration value and validates that it's
// a valid identifier
func getIdentifier(d *schema.ResourceData, key string) (string, diag.Diagnostics) {
	// We rely on Terraform to supply the correct types, and it's ok panic if this contract is broken
	s := d.Get(key).(string)
	if !identifierRegexp.MatchString(s) {
		return "", diag.Errorf("invalid value %q for identifier %q, validation expression is: %s", s, key, identifierRegexp.String())
	}

	return s, nil
}

func getClientSecret(d *schema.ResourceData) (string, diag.Diagnostics) {
	s := d.Get(azureClientSecretKey).(string)

	if s == "" {
		return "", diag.Errorf("empty client secret value for %q", azureClientSecretKey)
	}

	if len(s) > maxLengthClientSecret {
		return "", diag.Errorf(
			"invalid client secret value for %q, exceeds the maximum number of bytes allowed %d",
			azureClientSecretKey,
			maxLengthClientSecret,
		)
	}

	return s, nil
}
