// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/cyberark/idsec-sdk-golang/pkg/config"
	"github.com/cyberark/terraform-provider-idsec/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// GitCommit is the commit hash of the Idsec CLI application.
	GitCommit = "N/A"
	// BuildDate is the build date of the Idsec CLI application.
	BuildDate = "N/A"
	// Version is the version of the Idsec CLI application.
	Version = "N/A"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/cyberark/idsec",
		Debug:   debug,
	}
	config.SetIdsecToolInUse(config.IdsecToolTerraformProvider)
	config.GenerateCorrelationID()
	if debug || os.Getenv("TF_LOG") != "" {
		config.EnableVerboseLogging("DEBUG")
	}

	err := providerserver.Serve(context.Background(), provider.NewIdsecProvider(
		provider.IdsecProviderConfig{
			Version:   Version,
			GitCommit: GitCommit,
			BuildDate: BuildDate,
		},
	), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
