package csbmssqldbrunfailover

import (
	"context"

	"csbbrokerpakazure/providers/terraform-provider-csbmssqldbrunfailover/connector"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	resourceGroupKey = "resourceGroup"
	serverNameKey    = "serverName"
	failoverGroupKey = "failoverGroup"
)

func resourceRunFailover() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			resourceGroupKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			serverNameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			failoverGroupKey: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		CreateContext: createRunFailover,
		ReadContext:   nil,
		UpdateContext: nil,
		DeleteContext: nil,
		Description:   "Failover to the secondary database.",
	}
}

func createRunFailover(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var (
		resourceGroup,
		serverName,
		failoverGroup string
	)

	client := m.(*connector.Connector)

	for _, f := range []func() diag.Diagnostics{
		func() (diags diag.Diagnostics) {
			resourceGroup, diags = getIdentifier(d, resourceGroupKey)
			return
		},
		func() (diags diag.Diagnostics) {
			serverName, diags = getIdentifier(d, serverNameKey)
			return
		},
		func() (diags diag.Diagnostics) {
			failoverGroup, diags = getIdentifier(d, failoverGroupKey)
			return
		},
	} {
		if d := f(); d != nil {
			return d
		}
	}

	if err := client.CreateRunFailover(ctx, resourceGroup, serverName, failoverGroup); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
