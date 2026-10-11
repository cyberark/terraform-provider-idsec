// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package actions_test

import (
	"slices"
	"testing"

	"github.com/cyberark/terraform-provider-idsec/internal/actions"

	_ "github.com/cyberark/terraform-provider-idsec/internal/tfactions"
)

// TestEveryVersionBumpDeclaresAStateUpgrade asserts that a resource past its first schema version
// opts in to state upgrading by declaring at least one StateUpgrades entry for the immediately
// preceding version.
//
// Opting in (non-empty StateUpgrades) tells UpgradeState to register upgraders for every version
// in [0, ActionVersion): explicitly declared entries carry their Transform; undeclared versions are
// synthesized as pure passthroughs. A resource that skips opting in entirely keeps today's
// behaviour -- no upgraders registered -- which is Terraform's hard error for any practitioner
// holding older state. That error belongs in CI, on the commit that bumps the version.
func TestEveryVersionBumpDeclaresAStateUpgrade(t *testing.T) {
	allConfigs := actions.AllTerraformConfigs()

	if len(allConfigs) == 0 {
		t.Skip("No Terraform service configurations registered")
	}

	for _, config := range allConfigs {
		for _, resourceDef := range config.Resources {
			t.Run(config.ServiceName+"/"+resourceDef.ActionName, func(t *testing.T) {
				if resourceDef.ActionVersion <= 1 {
					// Version 0 and 1 are both a resource's first released schema -- there is no
					// earlier version anyone can hold state from.
					return
				}

				priorVersion := resourceDef.ActionVersion - 1
				if _, declared := resourceDef.StateUpgrades[priorVersion]; !declared {
					t.Errorf("resource %q in service %q is at ActionVersion %d but declares no StateUpgrades entry for version %d; "+
						"state written by the previous release fails to upgrade. Add StateUpgrades: map[int64]actions.StateUpgrade{%d: {}} "+
						"(an empty entry is a passthrough, which is right when the bump only made attributes write-only) or an entry with a "+
						"Transform when the bump restructured state. Undeclared intermediate versions are synthesized as passthroughs automatically.",
						resourceDef.ActionName, config.ServiceName, resourceDef.ActionVersion, priorVersion, priorVersion)
				}
			})
		}
	}
}

// TestStateUpgradeVersionsAreInRange asserts no resource declares an upgrade from a version
// Terraform can never ask for. A key at or above ActionVersion is dead code -- Terraform only
// requests an upgrade for state written below the current schema version -- and most often means a
// developer keyed the entry at the version they bumped TO rather than the one it upgrades FROM.
func TestStateUpgradeVersionsAreInRange(t *testing.T) {
	allConfigs := actions.AllTerraformConfigs()

	if len(allConfigs) == 0 {
		t.Skip("No Terraform service configurations registered")
	}

	for _, config := range allConfigs {
		for _, resourceDef := range config.Resources {
			t.Run(config.ServiceName+"/"+resourceDef.ActionName, func(t *testing.T) {
				for priorVersion := range resourceDef.StateUpgrades {
					if priorVersion < 0 || priorVersion >= resourceDef.ActionVersion {
						t.Errorf("resource %q in service %q declares a StateUpgrades entry for version %d, which is outside [0, %d); "+
							"entries are keyed by the version they upgrade FROM, so a bump to %d is keyed at %d",
							resourceDef.ActionName, config.ServiceName, priorVersion, resourceDef.ActionVersion,
							resourceDef.ActionVersion, resourceDef.ActionVersion-1)
					}
				}
			})
		}
	}
}

// TestPCloudAccountDeclaresTheWriteOnlySecretUpgrade pins the one resource this opt-in mechanism was
// introduced for: pcloud-account made its released "secret" write-only, which requires both the
// ActionVersion bump and a declared upgrade from the version before it. Without the pair, prior state
// hands the plain-text credential back and Terraform rejects the run with "Invalid resource state
// upgrade ... the provider returned a value for the write-only attribute".
func TestPCloudAccountDeclaresTheWriteOnlySecretUpgrade(t *testing.T) {
	var accountDef *actions.IdsecServiceTerraformResourceActionDefinition
	for _, config := range actions.AllTerraformConfigs() {
		for _, resourceDef := range config.Resources {
			if resourceDef.ActionName == "pcloud-account" {
				accountDef = resourceDef
			}
		}
	}

	if accountDef == nil {
		t.Skip("pcloud-account is not registered in this build")
	}

	if accountDef.ActionVersion < 2 {
		t.Fatalf("pcloud-account ActionVersion = %d, want at least 2: making the released \"secret\" write-only only takes effect from a bumped version",
			accountDef.ActionVersion)
	}
	if !slices.Contains(accountDef.WriteOnlyHashedAttributes, "secret") {
		t.Error(`pcloud-account no longer declares "secret" in WriteOnlyHashedAttributes; this test and the upgrade it pins assume it does`)
	}

	priorVersion := accountDef.ActionVersion - 1
	entry, declared := accountDef.StateUpgrades[priorVersion]
	if !declared {
		t.Fatalf("pcloud-account declares no StateUpgrades entry for version %d; state written before \"secret\" became write-only cannot upgrade",
			priorVersion)
	}
	if entry.Transform != nil {
		t.Errorf("pcloud-account's upgrade from version %d sets a Transform; the bump only made \"secret\" write-only, which needs a passthrough",
			priorVersion)
	}
}
