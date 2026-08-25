// Copyright CyberArk. 2026
// SPDX-License-Identifier: Apache-2.0

package schemas

import "testing"

// clearComputedAccount mirrors the relevant shape of TfIdsecCCEAWSAccount: a
// server-assigned computed "id" (the resource's read key) alongside other
// computed fields and a user-supplied, non-computed "services" list.
type clearComputedAccount struct {
	ID             string   `mapstructure:"id"`
	OnboardingType string   `mapstructure:"onboarding_type"`
	Status         string   `mapstructure:"status"`
	Services       []string `mapstructure:"services"`
}

// TestClearComputedAttributes_PreservesReadKeyID is the regression guard for the
// empty-ID-on-update bug (provider v0.5.0, PR #197 "OLY-17512"): update payloads
// strip computed attributes, keeping only the ImportID read key (the skip list).
// The CCE account resource lists "id" as computed, so if "id" is NOT in the skip
// list it gets wiped and the SDK issues `GET /api/aws/programmatic/account/`
// (empty id) -> 403. This locks in that the read key survives when skipped and
// that other computed attributes are still cleared.
func TestClearComputedAttributes_PreservesReadKeyID(t *testing.T) {
	computed := []string{"id", "onboarding_type", "status"}
	in := &clearComputedAccount{
		ID:             "abc123",
		OnboardingType: "terraform_provider",
		Status:         "Completely added",
		Services:       []string{"dpa"},
	}

	if err := ClearComputedAttributes(in, computed, []string{"id"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if in.ID != "abc123" {
		t.Fatalf("read key 'id' must be preserved when in the skip list, got %q", in.ID)
	}
	if in.OnboardingType != "" || in.Status != "" {
		t.Fatalf("non-key computed attributes must be cleared, got onboarding_type=%q status=%q",
			in.OnboardingType, in.Status)
	}
	if len(in.Services) != 1 || in.Services[0] != "dpa" {
		t.Fatalf("non-computed 'services' must be untouched, got %v", in.Services)
	}
}

// TestClearComputedAttributes_WipesReadKeyWithoutSkip documents the exact
// mechanism of the regression: with an empty skip list (i.e. no ImportID
// declared), the computed read key "id" is zeroed. This is what produced the
// empty-id GET -> 403 on update before ImportID was set on the CCE resources.
func TestClearComputedAttributes_WipesReadKeyWithoutSkip(t *testing.T) {
	computed := []string{"id", "onboarding_type", "status"}
	in := &clearComputedAccount{ID: "abc123", Status: "Completely added"}

	if err := ClearComputedAttributes(in, computed, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if in.ID != "" {
		t.Fatalf("without a skip list the computed 'id' is expected to be cleared "+
			"(regression scenario), got %q", in.ID)
	}
}

// clearComputedAzureSubscription mirrors the relevant shape of
// TfIdsecCCEAzureSubscription: a server-assigned computed "id" (the read key)
// plus the widest computed set among the CCE Azure resources, alongside a
// user-supplied, non-computed "subscription_id".
type clearComputedAzureSubscription struct {
	ID                  string `mapstructure:"id"`
	OnboardingType      string `mapstructure:"onboarding_type"`
	DisplayName         string `mapstructure:"display_name"`
	Status              string `mapstructure:"status"`
	ManagementGroupID   string `mapstructure:"management_group_id"`
	ManagementGroupName string `mapstructure:"management_group_name"`
	SubscriptionID      string `mapstructure:"subscription_id"`
}

// TestClearComputedAttributes_AzurePreservesReadKeyID is the Azure counterpart of
// TestClearComputedAttributes_PreservesReadKeyID. The CCE Azure resources also
// list "id" as computed, so the same update-payload stripping applies: the read
// key must survive when it is in the skip list while every other computed
// attribute is cleared and non-computed fields are left untouched.
func TestClearComputedAttributes_AzurePreservesReadKeyID(t *testing.T) {
	computed := []string{"id", "onboarding_type", "display_name", "status", "management_group_id", "management_group_name"}
	in := &clearComputedAzureSubscription{
		ID:                  "azsub123",
		OnboardingType:      "terraform_provider",
		DisplayName:         "prod-subscription",
		Status:              "Completely added",
		ManagementGroupID:   "mg-1",
		ManagementGroupName: "root-group",
		SubscriptionID:      "00000000-0000-0000-0000-000000000000",
	}

	if err := ClearComputedAttributes(in, computed, []string{"id"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if in.ID != "azsub123" {
		t.Fatalf("read key 'id' must be preserved when in the skip list, got %q", in.ID)
	}
	if in.OnboardingType != "" || in.DisplayName != "" || in.Status != "" ||
		in.ManagementGroupID != "" || in.ManagementGroupName != "" {
		t.Fatalf("non-key computed attributes must be cleared, got onboarding_type=%q display_name=%q status=%q management_group_id=%q management_group_name=%q",
			in.OnboardingType, in.DisplayName, in.Status, in.ManagementGroupID, in.ManagementGroupName)
	}
	if in.SubscriptionID != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("non-computed 'subscription_id' must be untouched, got %q", in.SubscriptionID)
	}
}

// TestClearComputedAttributes_AzureWipesReadKeyWithoutSkip documents the Azure
// regression scenario: with no ImportID (empty skip list) the computed read key
// "id" is zeroed, which is exactly what produced the empty-id GET -> 403 on
// update before ImportID was set on the CCE Azure resources.
func TestClearComputedAttributes_AzureWipesReadKeyWithoutSkip(t *testing.T) {
	computed := []string{"id", "onboarding_type", "display_name", "status", "management_group_id", "management_group_name"}
	in := &clearComputedAzureSubscription{ID: "azsub123", Status: "Completely added"}

	if err := ClearComputedAttributes(in, computed, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if in.ID != "" {
		t.Fatalf("without a skip list the computed 'id' is expected to be cleared "+
			"(regression scenario), got %q", in.ID)
	}
}
