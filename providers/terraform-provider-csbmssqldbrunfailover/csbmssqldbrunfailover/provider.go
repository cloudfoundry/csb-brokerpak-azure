// Package csbmssqldbrunfailover will contain the logic needed to create a provider to handle the failover
package csbmssqldbrunfailover

import (
	"context"

	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbmssqldbrunfailover/connector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	azureTenantIDKey       = "azure_tenant_id"
	azureClientIDKey       = "azure_client_id"
	azureClientSecretKey   = "azure_client_secret"
	azureSubscriptionIDKey = "azure_subscription_id"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			azureTenantIDKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			azureClientIDKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			azureClientSecretKey: {
				Type:     schema.TypeString,
				Required: true,
				// Set the Sensitive flag so output is hidden in the TF UI
				Sensitive: true,
			},
			azureSubscriptionIDKey: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		ConfigureContextFunc: configure,
		ResourcesMap: map[string]*schema.Resource{
			"csbmssqldbrunfailover_failover": resourceRunFailover(),
		},
	}
}

func configure(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var (
		azureTenantID,
		azureClientID,
		azureClientSecret,
		azureSubscriptionID string
	)

	for _, f := range []func() diag.Diagnostics{
		func() (diags diag.Diagnostics) {
			azureTenantID, diags = getIdentifier(d, azureTenantIDKey)
			return
		},
		func() (diags diag.Diagnostics) {
			azureClientID, diags = getIdentifier(d, azureClientIDKey)
			return
		},
		func() (diags diag.Diagnostics) {
			azureClientSecret, diags = getIdentifier(d, azureClientSecretKey)
			return
		},
		func() (diags diag.Diagnostics) {
			azureSubscriptionID, diags = getIdentifier(d, azureSubscriptionIDKey)
			return
		},
	} {
		if d := f(); d != nil {
			return nil, d
		}
	}

	return connector.NewConnector(azureTenantID, azureClientID, azureClientSecret, azureSubscriptionID), nil
}
