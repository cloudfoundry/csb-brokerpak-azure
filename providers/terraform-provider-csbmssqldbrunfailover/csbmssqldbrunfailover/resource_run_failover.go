package csbmssqldbrunfailover

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbmssqldbrunfailover/connector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	resourceGroupKey              = "resource_group"
	serverNameKey                 = "server_name"
	partnerServerNameKey          = "partner_server_name"
	failoverGroupKey              = "failover_group"
	partnerServerResourceGroupKey = "partner_server_resource_group"
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
			partnerServerNameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			failoverGroupKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			partnerServerResourceGroupKey: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
		CreateContext: create,
		ReadContext:   read,
		UpdateContext: update,
		DeleteContext: delete,
		Description:   "Failover to the secondary database.",
	}
}

func create(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var (
		partnerServerName,
		partnerServerResourceGroupName,
		failoverGroup string
	)

	client := m.(*connector.Connector)

	for _, f := range []func() diag.Diagnostics{
		func() (diags diag.Diagnostics) {
			_, diags = getIdentifier(d, resourceGroupKey)
			return
		},
		func() (diags diag.Diagnostics) {
			_, diags = getIdentifier(d, serverNameKey)
			return
		},
		func() (diags diag.Diagnostics) {
			partnerServerName, diags = getIdentifier(d, partnerServerNameKey)
			return
		},
		func() (diags diag.Diagnostics) {
			failoverGroup, diags = getIdentifier(d, failoverGroupKey)
			return
		},
		func() (diags diag.Diagnostics) {
			partnerServerResourceGroupName, diags = getIdentifier(d, partnerServerResourceGroupKey)
			return
		},
	} {
		if d := f(); d != nil {
			return d
		}
	}

	if err := client.RunFailover(ctx, partnerServerResourceGroupName, partnerServerName, failoverGroup); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(failoverGroup)

	return nil
}

func update(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return diag.FromErr(fmt.Errorf("update lifecycle not implemented"))
}

func read(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
			_, diags = getIdentifier(d, partnerServerNameKey)
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

	ok, err := client.ReadRunFailover(ctx, resourceGroup, serverName, failoverGroup)
	if err != nil {
		return diag.FromErr(err)
	}

	switch ok {
	case true:
		d.SetId(failoverGroup)
	default:
		d.SetId("")
	}

	return nil
}

func delete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
			_, diags = getIdentifier(d, partnerServerNameKey)
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

	if err := client.RunFailover(ctx, resourceGroup, serverName, failoverGroup); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
