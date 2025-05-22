package main

import (
	"context"
	"flag"
	"log"

	provider "terraform-provider-eci/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(
		&debug,
		"debug",
		false,
		"set to true to run the provider with support for debuggers like delve",
	)
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address:         "github.com/elice-dev/eci",
		Debug:           debug,
		ProtocolVersion: 6,
	}

	err := providerserver.Serve(context.Background(), provider.New(version, debug), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
