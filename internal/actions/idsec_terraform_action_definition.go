// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package actions

// IdsecServiceActionOperation defines the operation type for an Idsec service action, such as create, read, update, delete, or state.
type IdsecServiceActionOperation string

const (
	CreateOperation IdsecServiceActionOperation = "create"
	ReadOperation   IdsecServiceActionOperation = "read"
	UpdateOperation IdsecServiceActionOperation = "update"
	DeleteOperation IdsecServiceActionOperation = "delete"
	StateOperation  IdsecServiceActionOperation = "state"
)

// SingletonResourceImportDummyID is a constant used as a dummy ID for importing singleton resources in Terraform, where the resource does not have a natural unique identifier.
const SingletonResourceImportDummyID = "singleton"

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
	StateSchema             interface{}
	SensitiveAttributes     []string
	ExtraRequiredAttributes []string
	ComputedAsSetAttributes []string
	ImmutableAttributes     []string
	ComputedAttributes      []string
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
}

// IdsecServiceTerraformDataSourceActionDefinition is a struct that defines the structure of a data source action in the Idsec Terraform provider.
type IdsecServiceTerraformDataSourceActionDefinition struct {
	IdsecServiceBaseTerraformActionDefinition
	DataSourceAction string
}
