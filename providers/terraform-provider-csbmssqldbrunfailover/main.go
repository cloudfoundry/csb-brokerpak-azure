package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.gwd.broadcom.net/TNZ/csb-brokerpak-azure/terraform-provider-csbmssqldbrunfailover/csbmssqldbrunfailover"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: csbmssqldbrunfailover.Provider,
	})
}
