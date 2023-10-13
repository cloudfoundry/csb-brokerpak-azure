// Package csbsqlserver is a niche Terraform provider for Microsoft SQL Server
package csbsqlserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbsqlserver/connector"
)

const (
	serverKey       = "server"
	portKey         = "port"
	databaseKey     = "database"
	usernameKey     = "username"
	passwordKey     = "password"
	encryptKey      = "encrypt"
	ResourceNameKey = "csbsqlserver_binding"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema:               GetProviderSchema(),
		ConfigureContextFunc: ProviderContextFunc,
		ResourcesMap: map[string]*schema.Resource{
			ResourceNameKey: BindingResource(),
		},
	}
}

func GetProviderSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		usernameKey: {
			Type:     schema.TypeString,
			Required: true,
		},
		passwordKey: {
			Type:     schema.TypeString,
			Required: true,
		},
		encryptKey: {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func ProviderContextFunc(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
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
			username, diags = getServerIdentifier(d, usernameKey)
			return
		},
		func() (diags diag.Diagnostics) {
			password, diags = getServerPassword(d, passwordKey)
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
