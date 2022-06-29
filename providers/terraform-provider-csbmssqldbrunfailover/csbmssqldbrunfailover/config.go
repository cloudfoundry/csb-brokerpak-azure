package csbmssqldbrunfailover

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	identifierRegexp = regexp.MustCompile(`^[\w_.-]{1,64}$`)
	passwordRegexp   = regexp.MustCompile(`^[\w~_.-]{1,64}$`)
	validURL         = regexp.MustCompile(`^[\w.-]{1,253}$`)
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

// getIdentifierDefault gets an encrypt string configuration value and validates that it's
// valid. If no value is set, it returns the default value
func getEncrypt(d *schema.ResourceData, key string) (string, diag.Diagnostics) {
	// We rely on Terraform to supply the correct types, and it's ok panic if this contract is broken
	s := d.Get(key).(string)
	switch s {
	case "":
		return "true", nil // differs from driver default, but better to be secure by default
	case "false", "true", "disable":
		return s, nil
	default:
		return "", diag.Errorf("invalid value %q for %q, valid values are: true, false, disable", s, key)
	}
}

// getPassword gets a string configuration value and validates that it's
// a valid password
func getPassword(d *schema.ResourceData, key string) (string, diag.Diagnostics) {
	// We rely on Terraform to supply the correct types, and it's ok panic if this contract is broken
	s := d.Get(key).(string)
	if !passwordRegexp.MatchString(s) {
		return "", diag.Errorf("invalid password value for %q, validation expression is: %s", key, passwordRegexp.String())
	}

	return s, nil
}

// getPort gets a port configuration value and validates that it's
// a valid port
func getPort(d *schema.ResourceData, key string) (int, diag.Diagnostics) {
	// We rely on Terraform to supply the correct types, and it's ok panic if this contract is broken
	p := d.Get(key).(int)
	switch {
	case p <= 0:
		return 0, diag.Errorf("invalid port value %d for %q, port values must be positive integers", p, key)
	case p >= 65536:
		return 0, diag.Errorf("invalid port value %d for %q, port values must not exceed 16 bits", p, key)
	default:
		return p, nil
	}
}

// getURL gets a URL configuration value and validates that it's
// a valid URL
func getURL(d *schema.ResourceData, key string) (string, diag.Diagnostics) {
	// We rely on Terraform to supply the correct types, and it's ok panic if this contract is broken
	u := d.Get(key).(string)
	if !validURL.MatchString(u) {
		return "", diag.Errorf("invalid URL value %q for %q, validation expression is: %s", u, key, validURL.String())
	}

	return u, nil
}

// getRoles gets a configuration value and casts to a string slice
// We rely on Terraform to supply the correct types, and it's to panic if this contract is broken
func getRoles(d *schema.ResourceData, key string) ([]string, diag.Diagnostics) {
	var result []string
	// We rely on Terraform to supply the correct types, and it's ok panic if this contract is broken
	for i, e := range d.Get(key).([]any) {
		s := e.(string)
		if !identifierRegexp.MatchString(s) {
			return nil, diag.Errorf("invalid value %q for element %d of %q, validation expression is: %s", s, i, key, identifierRegexp.String())
		}

		result = append(result, e.(string))
	}

	return result, nil
}
