// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/services"
	"github.com/cyberark/terraform-provider-idsec/internal/featureadoption"
	"github.com/cyberark/terraform-provider-idsec/internal/schemas"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// IdsecServiceHelper provides common helper methods for working with service instances.
// This is embedded by both IdsecResource and IdsecDataSource.
type IdsecServiceHelper struct {
	serviceConfig *services.IdsecServiceConfig
	service       services.IdsecService
}

// getServiceNameTitled converts the service name to TitleCase format for reflection.
func (h *IdsecServiceHelper) getServiceNameTitled() string {
	serviceParts := strings.Split(h.serviceConfig.ServiceName, "-")
	titleCase := cases.Title(language.English)
	serviceNameTitled := ""
	for _, part := range serviceParts {
		serviceNameTitled += titleCase.String(part)
	}
	return strings.ReplaceAll(serviceNameTitled, "-", "")
}

// configureService retrieves and stores the service instance from the API.
// This should be called once during Configure() to set up the service.
// Returns an error if the service cannot be retrieved.
func (h *IdsecServiceHelper) configureService(idsecAPI *api.IdsecAPI) error {
	if idsecAPI == nil {
		return fmt.Errorf("idsecAPI is nil")
	}

	serviceNameTitled := h.getServiceNameTitled()

	// Try to get the service method using reflection
	serviceMethod, err := schemas.FindMethodByName(reflect.ValueOf(idsecAPI), serviceNameTitled)
	if err != nil || !serviceMethod.IsValid() {
		return fmt.Errorf("service method %s not found", serviceNameTitled)
	}

	// Call the service method to get the service instance
	serviceResults := serviceMethod.Call(nil)
	if len(serviceResults) < 2 {
		return fmt.Errorf("unexpected number of return values from service method")
	}

	// Check for error (second return value)
	if !serviceResults[1].IsNil() {
		err, ok := serviceResults[1].Interface().(error)
		if !ok {
			return fmt.Errorf("unexpected error type from service method")
		}
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Get the service (first return value)
	if !serviceResults[0].CanInterface() {
		return fmt.Errorf("cannot get service interface")
	}

	service, ok := serviceResults[0].Interface().(services.IdsecService)
	if !ok {
		return fmt.Errorf("service does not implement IdsecService interface")
	}

	// Validate that the service is actually usable by checking ServiceConfig
	if service == nil {
		return fmt.Errorf("service is nil")
	}

	// Try to get service config to verify service is properly initialized
	serviceConfig := service.ServiceConfig()
	if serviceConfig.ServiceName == "" {
		return fmt.Errorf("service not properly initialized - ServiceName is empty")
	}

	h.service = service
	return nil
}

// getServiceInstance retrieves the service instance.
// All services now implement the IdsecService interface which includes telemetry methods.
// Returns the service instance or nil if not configured.
func (h *IdsecServiceHelper) getServiceInstance() services.IdsecService {
	return h.service
}

// getService retrieves the service instance.
// All services now implement the IdsecService interface which includes telemetry methods.
// Returns the service instance or nil if not found.
func (h *IdsecServiceHelper) getService() services.IdsecService {
	return h.getServiceInstance()
}

// addTelemetryContextField adds a telemetry context field to the service and logs any errors.
func (h *IdsecServiceHelper) addTelemetryContextField(service services.IdsecService, name, shortName, value string) {
	if service == nil {
		return
	}
	if err := service.AddExtraContextField(name, shortName, value); err != nil {
		tflog.Warn(context.Background(), "Failed to add telemetry context field", map[string]interface{}{
			"field": name,
			"error": err.Error(),
		})
	}
}

// clearTelemetryContext clears all telemetry context fields from the service and logs any errors.
func (h *IdsecServiceHelper) clearTelemetryContext(service services.IdsecService) {
	if service == nil {
		return
	}
	if err := service.ClearExtraContext(); err != nil {
		tflog.Warn(context.Background(), "Failed to clear telemetry context", map[string]interface{}{"error": err.Error()})
	}
}

// getTerraformTypeName converts an action name to the Terraform resource/data source type name format.
// For example: "identity-role-admin-rights" becomes "idsec_identity_role_admin_rights".
func (h *IdsecServiceHelper) getTerraformTypeName(actionName string) string {
	return fmt.Sprintf("idsec_%s", strings.ReplaceAll(actionName, "-", "_"))
}

// buildFASTags builds the FAS report tags for a Terraform operation.
func (h *IdsecServiceHelper) buildFASTags(actionName, operation string) map[string]string {
	return map[string]string{
		featureadoption.TagKeyTFResource:  h.getTerraformTypeName(actionName),
		featureadoption.TagKeyTFOperation: operation,
		featureadoption.TagKeyTFService:   strings.Split(h.serviceConfig.ServiceName, "-")[0],
		featureadoption.TagKeyTFVersion:   providerVersion,
	}
}
