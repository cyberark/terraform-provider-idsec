// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type fakePrivateReader struct {
	data  map[string][]byte
	diags diag.Diagnostics
}

func (f fakePrivateReader) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if f.diags.HasError() {
		return nil, f.diags
	}
	return f.data[key], nil
}

func sortedPathKeys(paths map[string]bool) []string {
	keys := make([]string, 0, len(paths))
	for p := range paths {
		keys = append(keys, p)
	}
	sort.Strings(keys)
	return keys
}

func TestUserSetPaths(t *testing.T) {
	t.Parallel()

	t.Run("collect", func(t *testing.T) {
		t.Parallel()
		dataObj := types.ObjectValueMust(
			map[string]attr.Type{"connection_config": types.StringType, "other": types.StringType},
			map[string]attr.Value{"connection_config": types.StringValue("c"), "other": types.StringNull()},
		)
		paths := map[string]bool{}
		collectUserSetPaths(map[string]attr.Value{
			"secret_type": types.StringValue("password"),
			"omitted":     types.StringNull(),
			"data":        dataObj,
		}, "", paths)

		want := []string{"data", "data.connection_config", "secret_type"}
		if got := sortedPathKeys(paths); !reflect.DeepEqual(got, want) {
			t.Errorf("collectUserSetPaths = %v, want %v", got, want)
		}
	})

	t.Run("normalize", func(t *testing.T) {
		t.Parallel()
		if got, want := normalizeAttrPath("targets[0].name"), "targets.name"; got != want {
			t.Errorf("normalizeAttrPath = %q, want %q", got, want)
		}
	})
}

func TestUserSetHistoryPrivateState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// MarshalUserSetHistory wraps the paths in a self-describing envelope carrying the schema version,
	// a save timestamp, and the writing provider version. Even for empty paths it must emit a non-empty
	// blob (an empty key is treated as a delete by private state).
	blob, err := MarshalUserSetHistory(nil, "1.2.3")
	if err != nil {
		t.Fatalf("MarshalUserSetHistory(nil) error = %v", err)
	}
	var env userSetHistory
	if err := json.Unmarshal(blob, &env); err != nil {
		t.Fatalf("envelope unmarshal error = %v (blob=%q)", err, blob)
	}
	if env.SchemaVersion != UserSetHistorySchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", env.SchemaVersion, UserSetHistorySchemaVersion)
	}
	if env.ProviderVersion != "1.2.3" {
		t.Errorf("ProviderVersion = %q, want %q", env.ProviderVersion, "1.2.3")
	}
	if len(env.Paths) != 0 {
		t.Errorf("Paths = %v, want empty", env.Paths)
	}
	if _, err := time.Parse(time.RFC3339, env.SavedAt); err != nil {
		t.Errorf("SavedAt %q is not RFC3339: %v", env.SavedAt, err)
	}

	// A round-trip through the envelope format yields the membership set of paths.
	roundTrip, err := MarshalUserSetHistory([]string{"x", "y"}, "")
	if err != nil {
		t.Fatalf("MarshalUserSetHistory error = %v", err)
	}
	got := ReadUserSetPaths(ctx, fakePrivateReader{
		data: map[string][]byte{UserSetAttrsPrivateKey: roundTrip},
	})
	if !reflect.DeepEqual(got, map[string]bool{"x": true, "y": true}) {
		t.Errorf("ReadUserSetPaths(envelope) = %v", got)
	}

	// Legacy bare-array blobs written by older provider versions must still be readable.
	gotLegacy := ReadUserSetPaths(ctx, fakePrivateReader{
		data: map[string][]byte{UserSetAttrsPrivateKey: []byte(`["x","y"]`)},
	})
	if !reflect.DeepEqual(gotLegacy, map[string]bool{"x": true, "y": true}) {
		t.Errorf("ReadUserSetPaths(legacy array) = %v", gotLegacy)
	}

	if ReadUserSetPaths(ctx, nil) != nil {
		t.Error("nil reader should yield nil history")
	}
}

func TestClearRemovedAttributesHistoryGate(t *testing.T) {
	t.Parallel()

	type target struct {
		SecretType string `mapstructure:"secret_type"`
		HostName   string `mapstructure:"host_name"`
	}

	config := map[string]attr.Value{
		"secret_type": types.StringNull(),
		"host_name":   types.StringNull(),
	}
	state := map[string]attr.Value{
		"secret_type": types.StringValue("password"),
		"host_name":   types.StringValue("host-1"),
	}

	for _, tt := range []struct {
		name    string
		history map[string]bool
		want    target
	}{
		{"gated", map[string]bool{"secret_type": true}, target{HostName: "host-1"}},
		{"empty_history", map[string]bool{}, target{SecretType: "password", HostName: "host-1"}},
		{"nil_history", nil, target{SecretType: "password", HostName: "host-1"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tgt := target{SecretType: "password", HostName: "host-1"}
			clearRemovedAttributes(reflect.ValueOf(&tgt), config, state, nil, tt.history, "")
			if tgt != tt.want {
				t.Errorf("got %+v, want %+v", tgt, tt.want)
			}
		})
	}
}

func TestSyntheticHistoryFromState(t *testing.T) {
	t.Parallel()

	state := tfsdk.State{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"name":    schema.StringAttribute{},
				"enabled": schema.BoolAttribute{},
				"limit":   schema.Int64Attribute{},
				"tags":    schema.ListAttribute{ElementType: types.StringType},
				"metadata": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"policy_id":   schema.StringAttribute{},
						"description": schema.StringAttribute{},
					},
				},
			},
		},
		Raw: tftypes.NewValue(
			tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"enabled": tftypes.Bool,
					"limit":   tftypes.Number,
					"tags":    tftypes.List{ElementType: tftypes.String},
					"metadata": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"policy_id":   tftypes.String,
							"description": tftypes.String,
						},
					},
				},
			},
			map[string]tftypes.Value{
				"name":    tftypes.NewValue(tftypes.String, "policy"),
				"enabled": tftypes.NewValue(tftypes.Bool, false),
				"limit":   tftypes.NewValue(tftypes.Number, 0),
				"tags":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
				"metadata": tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"policy_id":   tftypes.String,
							"description": tftypes.String,
						},
					},
					map[string]tftypes.Value{
						"policy_id":   tftypes.NewValue(tftypes.String, "pid-1"),
						"description": tftypes.NewValue(tftypes.String, ""),
					},
				),
			},
		),
	}

	got, err := CollectStateSetPaths(context.Background(), &state)
	if err != nil {
		t.Fatalf("CollectStateSetPaths error = %v", err)
	}
	got = ReduceComputedPaths(got, []string{"metadata.status"}, []string{"metadata.policy_id"})
	want := []string{"metadata", "name"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Synthetic reduced paths = %v, want %v", got, want)
	}

	blob, err := MarshalSyntheticUserSetHistory([]string{"name"}, "1.0.0")
	if err != nil {
		t.Fatalf("MarshalSyntheticUserSetHistory error = %v", err)
	}
	var env userSetHistory
	if err := json.Unmarshal(blob, &env); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}
	if !env.Synthetic {
		t.Fatalf("Synthetic = false, want true")
	}
}
