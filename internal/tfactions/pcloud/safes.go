// Copyright CyberArk 2026
// SPDX-License-Identifier: Apache-2.0

package pcloud

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/safes/actions"
	safesmodels "github.com/cyberark/idsec-sdk-golang/pkg/services/pcloud/safes/models"
	tfactions "github.com/cyberark/terraform-provider-idsec/internal/actions"
)

// maxSafeMembersLimit is the maximum number of members allowed per Privilege Cloud safe.
const maxSafeMembersLimit = 64

// safeMemberPlanCounter tracks in-memory plan-time additions across all
// idsec_pcloud_safe_member resources in a single terraform plan invocation.
type safeMemberPlanCounter struct {
	mu          sync.Mutex
	remoteCache map[string]int // safeID -> current count from API (fetched once, then cached)
	delta       map[string]int // safeID -> members reserved by earlier plan calls this run
}

// newSafeMemberPlanCounter allocates a plan counter with initialized maps.
func newSafeMemberPlanCounter() *safeMemberPlanCounter {
	return &safeMemberPlanCounter{
		remoteCache: make(map[string]int),
		delta:       make(map[string]int),
	}
}

// checkAndReserve atomically checks whether adding one more member to the given safe
// would exceed the limit and, if not, reserves a slot for the current resource.
// When fetchRemote is nil the remote count is assumed to be 0 (safe does not exist
// yet so it has no members).
func (c *safeMemberPlanCounter) checkAndReserve(safeID string, fetchRemote func() (int, error)) (projected int, ok bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	remoteCount, cached := c.remoteCache[safeID]
	if !cached {
		if fetchRemote != nil {
			remoteCount, err = fetchRemote()
			if err != nil {
				return 0, false, err
			}
		}
		c.remoteCache[safeID] = remoteCount
	}

	projected = remoteCount + c.delta[safeID] + 1
	if projected > maxSafeMembersLimit {
		return projected, false, nil
	}
	c.delta[safeID]++
	return projected, true, nil
}

// safeMemberLimitValidator enforces that adding new safe member resources in a plan
// will not push the total membership count past maxSafeMembersLimit.
type safeMemberLimitValidator struct {
	counter *safeMemberPlanCounter
}

// ValidatePlan implements provider.IdsecPlanValidator.
func (v safeMemberLimitValidator) ValidatePlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	idsecAPI *api.IdsecAPI,
) {
	if !req.State.Raw.IsNull() {
		return
	}

	safeIDAttrPath := path.Root("safe_id")
	var safeID types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, safeIDAttrPath, &safeID)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if safeID.IsNull() {
		tflog.Debug(ctx, "safeMemberLimitValidator: safe_id is null, skipping count check")
		return
	}
	var id string
	var fetchRemote func() (int, error)
	if safeID.IsUnknown() {
		id = ""
		tflog.Debug(ctx, "safeMemberLimitValidator: safe_id is unknown (new safe), counting plan-time additions only")
	} else {
		id = safeID.ValueString()
		if id == "" {
			tflog.Debug(ctx, "safeMemberLimitValidator: safe_id is empty, skipping count check")
			return
		}
		fetchRemote = func() (int, error) {
			safesService, err := idsecAPI.PcloudSafes()
			if err != nil {
				return 0, fmt.Errorf("could not initialise pcloud safes service: %w", err)
			}
			stats, err := safesService.MembersStats(&safesmodels.IdsecPCloudGetSafeMembersStats{SafeID: id})
			if err != nil {
				return 0, fmt.Errorf("could not retrieve member count for safe %q: %w", id, err)
			}
			return stats.SafeMembersCount, nil
		}
	}

	projected, ok, err := v.counter.checkAndReserve(id, fetchRemote)
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("safeMemberLimitValidator: %s", err.Error()))
		return
	}

	if !ok {
		// Prefer safe_name for the error message; fall back to safe_id.
		safeLabel := id
		if safeLabel == "" {
			safeLabel = "(new safe)"
		} else {
			var safeName types.String
			if diags := req.Plan.GetAttribute(ctx, path.Root("safe_name"), &safeName); !diags.HasError() &&
				!safeName.IsUnknown() && !safeName.IsNull() && safeName.ValueString() != "" {
				safeLabel = safeName.ValueString()
			}
		}
		resp.Diagnostics.AddAttributeError(
			safeIDAttrPath,
			"Safe member limit exceeded",
			fmt.Sprintf(
				"Adding this member would bring safe %q to %d member(s), which exceeds the maximum of %d. "+
					"Remove an existing member or reduce the number of members being added in this plan.",
				safeLabel, projected, maxSafeMembersLimit,
			),
		)
	}
}

func init() {
	safeMemberCounter := newSafeMemberPlanCounter()

	_ = tfactions.Register(tfactions.TerraformServiceConfig{
		ServiceName: "pcloud-safes",
		Resources: []*tfactions.IdsecServiceTerraformResourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-safe", ActionDescription: "Privilege Cloud Safe resource, manages Privilege Cloud Safes information and metadata.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ComputedAttributes: []string{
						"safe_number",
						"creator",
						"creation_time",
						"last_modification_time",
						"is_expired_member",
					},
					ImmutableAttributes: []string{
						"auto_purge_enabled",
					},
					StateSchema: &safesmodels.IdsecPCloudSafe{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "create", tfactions.ReadOperation: "get", tfactions.UpdateOperation: "update", tfactions.DeleteOperation: "delete"},
				ImportID:            "safe_id",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-safe-member", ActionDescription: "Privilege Cloud safe member resource, manages Privilege Cloud Safe members and their relevant permissions.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ComputedAttributes: []string{
						"member_id",
						"safe_number",
						"safe_name",
						"is_expired_membership_enabled",
						"is_predefined_user",
						"is_read_only",
					},
					ImmutableAttributes: []string{
						"search_in",
						"member_name",
						"member_type",
						"safe_id",
					},
					StateSchema: &safesmodels.IdsecPCloudSafeMember{},
				},
				SupportedOperations: []tfactions.IdsecServiceActionOperation{tfactions.CreateOperation, tfactions.ReadOperation, tfactions.UpdateOperation, tfactions.DeleteOperation, tfactions.StateOperation},
				ActionsMappings:     map[tfactions.IdsecServiceActionOperation]string{tfactions.CreateOperation: "add-member", tfactions.ReadOperation: "get-member", tfactions.UpdateOperation: "update-member", tfactions.DeleteOperation: "delete-member"},
				ImportID:            "safe_id:member_name",
				PlanValidators:      []tfactions.IdsecPlanValidator{safeMemberLimitValidator{counter: safeMemberCounter}},
			},
		},
		DataSources: []*tfactions.IdsecServiceTerraformDataSourceActionDefinition{
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-safe", ActionDescription: "Privilege Cloud Safe data source, reads safe information and metadata, based on the Safe ID.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"safe_id"},
					StateSchema:             &safesmodels.IdsecPCloudSafe{},
				},
				DataSourceAction: "get",
			},
			{
				IdsecServiceBaseTerraformActionDefinition: tfactions.IdsecServiceBaseTerraformActionDefinition{
					IdsecServiceBaseActionDefinition: tfactions.IdsecServiceBaseActionDefinition{
						ActionName: "pcloud-safe-member", ActionDescription: "Privilege Cloud Safe Member data source, reads Safe member information and metadata, based on the Safe ID and the member name.", ActionVersion: 1, Schemas: actions.ActionToSchemaMap,
					},
					ExtraRequiredAttributes: []string{"safe_id", "member_name"},
					StateSchema:             &safesmodels.IdsecPCloudSafeMember{},
				},
				DataSourceAction: "get-member",
			},
		},
	})
}
