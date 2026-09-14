// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package actions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
)

// IdsecServiceActionOperation defines the operation type for an Idsec service action, such as create, read, update, delete, or state.
type IdsecServiceActionOperation string

// IdsecPlanValidator is the interface that plan-time API validators that are more complex than simple attribute checks.
//
// Implementors are registered on IdsecServiceTerraformResourceActionDefinition.PlanValidators
// and invoked during ModifyPlan, after the provider has been configured and the IdsecAPI
// client is ready. Add error diagnostics to resp to block the apply.
//
// Example implementation:
//
//	type safeMemberLimitValidator struct{}
//
//	func (v safeMemberLimitValidator) ValidatePlan(
//	    ctx context.Context,
//	    req resource.ModifyPlanRequest,
//	    resp *resource.ModifyPlanResponse,
//	    api *idsecapi.IdsecAPI,
//	) {
//	    // only block on create (state is null)
//	    if !req.State.Raw.IsNull() { return }
//	    // ... API call, count members, add error if over limit
//	}
type IdsecPlanValidator interface {
	ValidatePlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, api *api.IdsecAPI)
}

const (
	CreateOperation IdsecServiceActionOperation = "create"
	ReadOperation   IdsecServiceActionOperation = "read"
	UpdateOperation IdsecServiceActionOperation = "update"
	DeleteOperation IdsecServiceActionOperation = "delete"
	StateOperation  IdsecServiceActionOperation = "state"
)

// SingletonResourceImportDummyID is a constant used as a dummy ID for importing singleton resources in Terraform, where the resource does not have a natural unique identifier.
const SingletonResourceImportDummyID = "singleton"

// DocNoteSeverity defines the severity of a custom documentation note.
//
// The severity controls how the note is rendered per documentation target:
//   - DocNoteInfo    -> MkDocs "note"    / Terraform Registry "->"
//   - DocNoteWarning -> MkDocs "warning" / Terraform Registry "~>"
//   - DocNoteDanger  -> MkDocs "danger"  / Terraform Registry "!>"
type DocNoteSeverity string

const (
	DocNoteInfo    DocNoteSeverity = "info"
	DocNoteWarning DocNoteSeverity = "warning"
	DocNoteDanger  DocNoteSeverity = "danger"
)

// DocNote is a single custom documentation note injected into the generated
// resource or data source documentation.
type DocNote struct {
	Severity       DocNoteSeverity
	Body           string
	BreakingChange bool
}

// IdsecServiceBaseActionDefinition is a struct that defines the base structure of a Terraform action definition.
type IdsecServiceBaseActionDefinition struct {
	ActionName        string
	Enabled           *bool
	ActionDescription string
	ActionVersion     int64
	Schemas           map[string]interface{}
}

// ActionDefinitionName returns the name of the action definition.
func (a *IdsecServiceBaseActionDefinition) ActionDefinitionName() string {
	return a.ActionName
}

// IsEnabled returns whether the action is enabled for registration.
// Returns true if Enabled is nil (default) or explicitly set to true.
func (a *IdsecServiceBaseActionDefinition) IsEnabled() bool {
	return a.Enabled == nil || *a.Enabled
}

// IdsecServiceBaseTerraformActionDefinition is a struct that defines the structure of an action in the Idsec Terraform provider.
type IdsecServiceBaseTerraformActionDefinition struct {
	IdsecServiceBaseActionDefinition
	StateSchema               interface{}
	SensitiveAttributes       []string
	ExtraRequiredAttributes   []string
	ComputedAsSetAttributes   []string
	ImmutableAttributes       []string
	ComputedAttributes        []string
	HistoryComputedAttributes []string
	// SemanticEqualityAttributes maps an attribute's field name to the semantic-equality plan
	// modifier kind that should be attached (e.g. schemas.SemanticEqualityCaseInsensitive,
	// schemas.SemanticEqualityTrailingSlash). See schemas.SemanticEqualityKind.
	SemanticEqualityAttributes map[string]schemas.SemanticEqualityKind
	PageNotes                  []DocNote
	AttributeNotes             map[string][]DocNote
}

// IdsecServiceTerraformResourceActionDefinition is a struct that defines the structure of a resource action in the Idsec Terraform provider.
type IdsecServiceTerraformResourceActionDefinition struct {
	IdsecServiceBaseTerraformActionDefinition
	RawStateInference   bool
	ReadSchemaPath      string
	DeleteSchemaPath    string
	SupportedOperations []IdsecServiceActionOperation
	ActionsMappings     map[IdsecServiceActionOperation]string
	ImportID            string
	PlanValidators      []IdsecPlanValidator
	WriteOnlyAttributes map[string]string
	// WriteOnlyHashedAttributes lists top-level scalar attribute paths that desugar into
	// write-only mode with an automatically synthesized "<attr>_write_only_hash" trigger, instead
	// of a manually declared one: the schema generator marks the attribute write-only and adds the
	// hash sibling itself, so no entry for it belongs in WriteOnlyAttributes as well.
	WriteOnlyHashedAttributes []string
	// DeprecationMessage is shown as a Terraform warning whenever this resource
	// appears in a plan. Leave empty for non-deprecated resources.
	DeprecationMessage string
}

// IdsecServiceTerraformDataSourceActionDefinition is a struct that defines the structure of a data source action in the Idsec Terraform provider.
type IdsecServiceTerraformDataSourceActionDefinition struct {
	IdsecServiceBaseTerraformActionDefinition
	DataSourceAction string
}
