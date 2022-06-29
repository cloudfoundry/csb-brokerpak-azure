package csbmssqldbrunfailover

import (
"context"
	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/connector"

	//"csbbrokerpakazure/providers/terraform-provider-csbmssqldbrunfailover/connector"

"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	azureTenantIDKey = "azureTenantID"
	azureClientIDKey     = "azureClientID"
	azureClientSecretKey   = "azureClientSecret"
	azureSubscriptionIDKey = "azureSubscriptionID"
	resourceGroupKey    = "resourceGroup"
	serverNameKey          = "serverName"
	failoverGroupKey = "failoverGroup"
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
			//resourceGroupKey: {
			//	Type:     schema.TypeString,
			//	Required: true,
			//},
			//serverNameKey: {
			//	Type:     schema.TypeString,
			//	Required: true,
			//},
			//failoverGroupKey: {
			//	Type:     schema.TypeString,
			//	Required: true,
			//},
		},
		ConfigureContextFunc: configure,
		ResourcesMap: map[string]*schema.Resource{
			"csbmssqldbrunfailover": runFailoverResource(),
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
			azureClientSecret, diags = getIdentifier(d, azureClientSecretKey)
			return
		},
		func() (diags diag.Diagnostics) {
				azureSubscriptionID, diags = getIdentifier(d, azureSubscriptionIDKey)
			return
		},
		//func() (diags diag.Diagnostics) {
		//	serverName, diags = getIdentifier(d, serverNameKey)
		//	return
		//},
		//func() (diags diag.Diagnostics) {
		//		failoverGroup, diags = getIdentifier(d, failoverGroupKey)
		//	return
		//},
		//func() (diags diag.Diagnostics) {
		//		resourceGroup, diags = getIdentifier(d, resourceGroupKey)
		//	return
		//},

	} {
		if d := f(); d != nil {
			return nil, d
		}
	}

	return connector.New(azureTenantID, azureClientID, azureClientSecret, azureSubscriptionID), nil
}

