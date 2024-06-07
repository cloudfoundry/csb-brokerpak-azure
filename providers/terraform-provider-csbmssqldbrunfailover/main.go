package main

import (
	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbmssqldbrunfailover/csbmssqldbrunfailover"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbmssqldbrunfailover.Provider,
	})
}
