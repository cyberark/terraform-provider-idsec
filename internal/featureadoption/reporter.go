// Copyright (c) CyberArk.
// SPDX-License-Identifier: Apache-2.0

package featureadoption

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	api "github.com/cyberark/idsec-sdk-golang/pkg"
	"github.com/cyberark/idsec-sdk-golang/pkg/config"
	"github.com/cyberark/idsec-sdk-golang/pkg/featureadoption"
)

const (
	// MetricKey is the FAS metric key for the Terraform provider.
	MetricKey = "IDSGO.idsec_terraform_provider.usage"
	// TagKeyTFService is the FAS tag key for the Terraform service context.
	TagKeyTFService = "tfs"
	// TagKeyTFOperation is the FAS tag key for the Terraform operation context.
	TagKeyTFOperation = "tfo"
	// TagKeyTFVersion is the FAS tag key for the provider version.
	TagKeyTFVersion = "tfv"
	// TagKeyTFResource is the FAS tag key for the Terraform resource type name.
	TagKeyTFResource = "tfr"
)

// ReportOptions holds optional parameters for FAS reporting. Extensible for future tags (e.g. operation_duration_ms).
type ReportOptions struct {
	// OperationDuration is how long the operation took. If set, adds "time" to custom_data (milliseconds).
	OperationDuration *time.Duration
	// OperationStatus is the outcome of the operation ("success" or "failure").
	// Reported as "ops" in custom_data.
	OperationStatus string
	// Message is the diagnostic output of the operation.
	// On failure it contains error summaries; on success it may be empty.
	// Reported as "message" in custom_data.
	Message string
	// ExtraTags adds custom tags to the report. Keys must match FAS pattern ^[a-zA-Z0-9_]+$.
	ExtraTags map[string]string
}

// ReportOperationDefer returns a function to be used with defer. Call it at the start of an operation;
// the returned function will run on return and report the operation status, message, and duration to FAS.
//
// Parameters:
//   - ctx: Context for logging
//   - idsecAPI: The API instance for auth token resolution
//   - diagnostics: Pointer to the response diagnostics; inspected at defer time to determine status and message
//   - extraTags: Additional tags to include in the FAS report (e.g. tfo, tfr, tfs, tfv). May be nil.
func ReportOperationDefer(ctx context.Context, idsecAPI *api.IdsecAPI, diagnostics *diag.Diagnostics, extraTags map[string]string) func() {
	start := time.Now()
	return func() {
		dur := time.Since(start)
		status := "success"
		message := ""
		if diagnostics.HasError() {
			status = "failure"
			message = diagErrorSummaries(diagnostics)
		}
		ReportAsync(ctx, idsecAPI, &ReportOptions{
			OperationDuration: &dur,
			OperationStatus:   status,
			Message:           message,
			ExtraTags:         extraTags,
		})
	}
}

// diagErrorSummaries returns a semicolon-separated string of all error diagnostic summaries.
func diagErrorSummaries(diagnostics *diag.Diagnostics) string {
	var summaries []string
	for _, d := range diagnostics.Errors() {
		if s := d.Summary(); s != "" {
			summaries = append(summaries, s)
		}
	}
	return strings.Join(summaries, "; ")
}

// ReportAsync sends a feature adoption report to FAS (synchronous so logs are emitted before handler returns).
// The token is resolved from idsecAPI internally by the SDK. Skips if FAS URL is unset, telemetry is disabled,
// no auth configured, or token is not JWT (PVWA).
// For resources use isDataSource=false (adds terraform_resource tag); for data sources use isDataSource=true (adds terraform_data_source tag).
// opts can be nil; when provided, OperationDuration/OperationStatus/Message are merged into custom_data
// and ExtraTags are merged into report tags.
func ReportAsync(ctx context.Context, idsecAPI *api.IdsecAPI, opts *ReportOptions) {
	tags := buildTags(opts)
	customData := buildCustomData(opts)

	reportCtx := context.Background()
	reportOpts := &featureadoption.ReportOpts{
		CustomData: customData,
	}
	msg, err := featureadoption.ReportWithAPI(reportCtx, idsecAPI, MetricKey, tags, reportOpts)
	if err != nil {
		tflog.Warn(ctx, "FAS report failed", map[string]interface{}{
			"error": err.Error(),
		})
	} else if msg != "" {
		tflog.Debug(ctx, msg, map[string]interface{}{
			"metric_key": MetricKey,
		})
	}
}

func buildTags(opts *ReportOptions) (tags map[string]string) {
	tags = make(map[string]string)
	if opts != nil {
		for k, v := range opts.ExtraTags {
			if k != "" && v != "" {
				tags[k] = v
			}
		}
	}
	return tags
}

func buildCustomData(opts *ReportOptions) (customData map[string]interface{}) {
	customData = make(map[string]interface{})
	customData["correlation_id"] = config.CorrelationID()

	if opts == nil {
		return customData
	}
	if opts.OperationDuration != nil {
		customData["duration"] = fmt.Sprintf("%d", opts.OperationDuration.Milliseconds())
	}
	if opts.OperationStatus != "" {
		customData["ops"] = opts.OperationStatus
	}
	if opts.Message != "" {
		customData["message"] = opts.Message
	}
	return customData
}
