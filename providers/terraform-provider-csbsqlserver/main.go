package main

import (
	"csbbrokerpakazure/providers/terraform-provider-csbsqlserver/csbsqlserver"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbsqlserver.Provider,
	})
}
