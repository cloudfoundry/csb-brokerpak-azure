// Package csbsqlserver is a niche Terraform provider for Microsoft SQL Server
package csbsqlserver

import (
	"context"

	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/connector"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	serverKey           = "server"
	portKey             = "port"
	databaseKey         = "database"
	providerUsernameKey = "username"
	providerPasswordKey = "password"
	encryptKey          = "encrypt"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			serverKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			portKey: {
				Type:     schema.TypeInt,
				Required: true,
			},
			databaseKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			providerUsernameKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			providerPasswordKey: {
				Type:     schema.TypeString,
				Required: true,
			},
			encryptKey: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ConfigureContextFunc: configure,
		ResourcesMap: map[string]*schema.Resource{
			"csbsqlserver_binding": bindingResource(),
		},
	}
}

func configure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	var (
		server   string
		port     int
		username string
		password string
		database string
		encrypt  string
	)

	for _, f := range []func() diag.Diagnostics{
		func() (diags diag.Diagnostics) {
			server, diags = getURL(d, serverKey)
			return
		},
		func() (diags diag.Diagnostics) {
			port, diags = getPort(d, portKey)
			return
		},
		func() (diags diag.Diagnostics) {
			username, diags = getServerIdentifier(d, providerUsernameKey)
			return
		},
		func() (diags diag.Diagnostics) {
			password, diags = getServerPassword(d, providerPasswordKey)
			return
		},
		func() (diags diag.Diagnostics) {
			database, diags = getServerIdentifier(d, databaseKey)
			return
		},
		func() (diags diag.Diagnostics) {
			encrypt, diags = getEncrypt(d, encryptKey)
			return
		},
	} {
		if d := f(); d != nil {
			return nil, d
		}
	}

	return connector.New(server, port, username, password, database, encrypt), nil
}
