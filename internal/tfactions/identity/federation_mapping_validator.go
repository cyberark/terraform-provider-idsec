// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/services/identity/federations"
	federationsmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/identity/federations/models"

	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// federationMappingValidator reports at plan time the user mapping, group mapping, attribute mapping
// and routing settings the server rejects, or accepts and silently rewrites, which would otherwise
// fail the apply or show up as drift after it.
//
// The rules are the SDK's: Create and Update check the settings they would store with
// federations.ValidateCreateFederationSettings and ValidateUpdateFederationSettings, and this calls
// the same functions on the request the apply would send, so the CLI, the SDK and Terraform plans
// get the same answer. On create the request is the configuration, with the SDK applying the
// server defaults to what it leaves unset; on update it is the plan, over the prior state. A
// problem depending on a value the configuration only knows at apply is left for the apply.
//
// The one rule of its own is Terraform's: an empty user_agents is stored as ["Any"], which is fine
// for the server but reads back as drift.
type federationMappingValidator struct{}

// ValidatePlan implements tfactions.IdsecPlanValidator.
func (v federationMappingValidator) ValidatePlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	_ *api.IdsecAPI,
) {
	if req.Plan.Raw.IsNull() {
		return
	}
	v.checkEmptyUserAgents(ctx, req, resp)

	err, decodeErr := v.settingsError(ctx, req)
	if decodeErr != nil {
		tflog.Debug(ctx, fmt.Sprintf("federationMappingValidator: skipping the settings check: %s", decodeErr.Error()))
		return
	}
	unknown := unknownAttributes(req.Config.Raw)
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		return
	}
	for _, e := range joined.Unwrap() {
		var problem *federations.FederationSettingsError
		if !errors.As(e, &problem) || slices.ContainsFunc(problem.DependsOn, func(attribute string) bool { return unknown[attribute] }) {
			continue
		}
		resp.Diagnostics.AddAttributeError(problemPath(problem), problem.Summary, problem.Detail)
	}
}

// settingsError runs the SDK check the apply would run. The second error is a decoding failure, on
// which the check is skipped rather than guessed at.
func (v federationMappingValidator) settingsError(ctx context.Context, req resource.ModifyPlanRequest) (error, error) {
	if req.State.Raw.IsNull() {
		decoded, err := schemas.StructFromConfigObject(ctx, &req.Config, federationsmodels.IdsecIdentityCreateFederation{})
		if err != nil {
			return nil, err
		}
		create, ok := decoded.(federationsmodels.IdsecIdentityCreateFederation)
		if !ok {
			return nil, fmt.Errorf("decoded the configuration as %T", decoded)
		}
		return federations.ValidateCreateFederationSettings(&create), nil
	}
	decodedState, err := schemas.StructFromStateObject(ctx, &req.State, federationsmodels.IdsecIdentityFederation{})
	if err != nil {
		return nil, err
	}
	stored, ok := decodedState.(federationsmodels.IdsecIdentityFederation)
	if !ok {
		return nil, fmt.Errorf("decoded the state as %T", decodedState)
	}
	decodedPlan, err := schemas.StructFromPlanObject(ctx, &req.Plan, federationsmodels.IdsecIdentityUpdateFederation{})
	if err != nil {
		return nil, err
	}
	update, ok := decodedPlan.(*federationsmodels.IdsecIdentityUpdateFederation)
	if !ok {
		return nil, fmt.Errorf("decoded the plan as %T", decodedPlan)
	}
	return federations.ValidateUpdateFederationSettings(&stored, update), nil
}

// checkEmptyUserAgents rejects an empty user_agents, which the server stores as ["Any"], so the
// set would read back different from the configuration.
func (v federationMappingValidator) checkEmptyUserAgents(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var set types.Set
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("user_agents"), &set)...)
	if !set.IsNull() && !set.IsUnknown() && len(set.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(path.Root("user_agents"), "No User Agent Selected",
			`An empty user_agents is stored as ["Any"]: omit it, or set it to ["Any"], instead.`)
	}
}

// problemPath is the attribute path a settings problem is reported on.
func problemPath(problem *federations.FederationSettingsError) path.Path {
	at := path.Root(problem.Attribute)
	if problem.Index >= 0 {
		at = at.AtListIndex(problem.Index)
	}
	if problem.Field != "" {
		at = at.AtName(problem.Field)
	}
	return at
}

// unknownAttributes returns the top-level attributes of a configuration that are unknown, or hold
// an unknown value, until apply.
func unknownAttributes(config tftypes.Value) map[string]bool {
	unknown := map[string]bool{}
	attributes := map[string]tftypes.Value{}
	if config.IsNull() || !config.IsKnown() || config.As(&attributes) != nil {
		return unknown
	}
	for name, value := range attributes {
		if !value.IsFullyKnown() {
			unknown[name] = true
		}
	}
	return unknown
}
