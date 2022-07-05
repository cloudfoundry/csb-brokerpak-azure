package csbmssqldbrunfailover

import (
	"context"

	"csbbrokerpakazure/providers/terraform-provider-csbmssqldbrunfailover/connector"

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

func configure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
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
			azureClientSecret, diags = getClientSecret(d)
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
