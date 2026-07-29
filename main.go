package main

import (
	"context"
	"flag"
	"log"

	"github.com/AdamBowers/terraform-provider-pve/v2/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var debug bool
	var address string

	flag.BoolVar(&debug, "debug", false, "run provider with debugger support.")
	flag.StringVar(&address, "address", "hashicorp.com/edu/pve", "registry address.")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: address,
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
