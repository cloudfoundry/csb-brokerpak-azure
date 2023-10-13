package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/cloudfoundry/csb-brokerpak-azure/terraform-provider-csbsqlserver/csbsqlserver"
)

func main() {
	var debug bool

	// It is important to start a provider in debug mode only when you intend to debug it, as its behavior will
	// change in minor ways from normal operation of providers.
	//
	// The main differences are:
	//
	// * Terraform will not start the provider process; it must be run manually.
	// * The provider will no longer be restarted once per walk of the Terraform graph;
	//   instead the same provider process will be reused until the command is completed.
	//
	// Note: We need to disable compiler optimization and inlining to have the debugger
	// work efficiently with the provider binary.
	// To do so, build the provider binary with the necessary flags: go build -gcflags="all=-N -l" -o xxxx
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debug,
		ProviderFunc: csbsqlserver.Provider,
	}

	if debug {
		// see tf configuration in examples folder
		opts.ProviderAddr = "cloudfoundry/csbsqlserver"
	}

	plugin.Serve(opts)
}
