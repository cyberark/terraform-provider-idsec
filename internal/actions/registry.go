// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package actions

import "fmt"

// TerraformServiceConfig holds the Terraform-specific configuration for a service,
// including its resources and data sources.
type TerraformServiceConfig struct {
	ServiceName string
	Resources   []*IdsecServiceTerraformResourceActionDefinition
	DataSources []*IdsecServiceTerraformDataSourceActionDefinition
}

var terraformRegistry []TerraformServiceConfig

// Register adds a TerraformServiceConfig to the registry.
func Register(config TerraformServiceConfig) error {
	for _, existing := range terraformRegistry {
		if existing.ServiceName == config.ServiceName {
			return fmt.Errorf("terraform service config %s already registered", config.ServiceName)
		}
	}

	if isEnableAttributeActive() {
		config = filterEnabledActions(config)
	}

	terraformRegistry = append(terraformRegistry, config)
	return nil
}

// AllTerraformConfigs returns all registered Terraform service configurations.
func AllTerraformConfigs() []TerraformServiceConfig {
	configs := make([]TerraformServiceConfig, len(terraformRegistry))
	copy(configs, terraformRegistry)
	return configs
}

// releasedFeaturesOnly controls whether Enable attribute filtering is applied.
// Set via ldflags: -ldflags "-X github.com/cyberark/terraform-provider-idsec/internal/actions.releasedFeaturesOnly=true".
var releasedFeaturesOnly = "false"

func isEnableAttributeActive() bool {
	return releasedFeaturesOnly == "true"
}

func filterEnabledActions(config TerraformServiceConfig) TerraformServiceConfig {
	filtered := TerraformServiceConfig{
		ServiceName: config.ServiceName,
	}

	for _, r := range config.Resources {
		if r.IsEnabled() {
			filtered.Resources = append(filtered.Resources, r)
		}
	}
	for _, d := range config.DataSources {
		if d.IsEnabled() {
			filtered.DataSources = append(filtered.DataSources, d)
		}
	}

	return filtered
}
