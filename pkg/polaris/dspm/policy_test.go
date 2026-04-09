// Copyright 2026 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package dspm

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/dspm"
)

func TestPolicyManagement(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	dspmClient := Wrap(client)

	// Discover valid filter values for the sensitivity filter type so the
	// test uses values the API actually accepts.
	sensMeta, err := dspmClient.FilterValues(ctx, dspm.FilterTypeDocumentSensitivity, "")
	if err != nil {
		t.Fatalf("failed to discover sensitivity filter values: %v", err)
	}
	if len(sensMeta.Values) == 0 {
		t.Fatal("no values returned for sensitivity filter")
	}
	sensValue := sensMeta.Values[0].ID

	exposureMeta, err := dspmClient.FilterValues(ctx, dspm.FilterTypeDocumentExposure, "")
	if err != nil {
		t.Fatalf("failed to discover exposure filter values: %v", err)
	}
	if len(exposureMeta.Values) == 0 {
		t.Fatal("no values returned for exposure filter")
	}
	exposureValue := exposureMeta.Values[0].ID

	// Create a data security policy with a complex nested filter and a
	// threshold filter:
	//   filter: AND(
	//     sensitivity IS <value>,
	//     OR(
	//       exposure IS <value>,
	//       sensitivity IS_NOT <value>,
	//     ),
	//   )
	//   thresholdFilter: AND(exposure IS <value>)
	thresholdFilter := dspm.GroupConfig{
		Op: dspm.LogicalAnd,
		Filters: []dspm.Node{
			{
				Config: &dspm.Config{
					Type:         dspm.FilterTypeDocumentExposure,
					Values:       []string{exposureValue},
					Relationship: dspm.RelIs,
				},
			},
		},
	}
	policyID, err := dspmClient.CreatePolicy(ctx, dspm.CreateInput{
		Name:        "SDK Integration Test Policy",
		Description: "Created by SDK integration test",
		Category:    dspm.CategoryOverexposed,
		Severity:    dspm.SeverityHigh,
		Filter: dspm.GroupConfig{
			Op: dspm.LogicalAnd,
			Filters: []dspm.Node{
				{
					Config: &dspm.Config{
						Type:         dspm.FilterTypeDocumentSensitivity,
						Values:       []string{sensValue},
						Relationship: dspm.RelIs,
					},
				},
				{
					GroupConfig: &dspm.GroupConfig{
						Op: dspm.LogicalOr,
						Filters: []dspm.Node{
							{
								Config: &dspm.Config{
									Type:         dspm.FilterTypeDocumentExposure,
									Values:       []string{exposureValue},
									Relationship: dspm.RelIs,
								},
							},
							{
								Config: &dspm.Config{
									Type:         dspm.FilterTypeDocumentSensitivity,
									Values:       []string{sensValue},
									Relationship: dspm.RelIsNot,
								},
							},
						},
					},
				},
			},
		},
		ThresholdFilter: &thresholdFilter,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Ensure cleanup on test failure. Registered after creation success so
	// that a failed create does not attempt to delete a zero-value ID.
	t.Cleanup(func() {
		if err := dspmClient.DeletePolicy(ctx, policyID); err != nil {
			t.Errorf("failed to delete policy: %v", err)
		}
	})

	// Get policy by ID and verify all fields.
	policy, err := dspmClient.PolicyByID(ctx, policyID)
	if err != nil {
		t.Fatal(err)
	}
	if policy.ID != policyID {
		t.Errorf("ID: got %q, want %q", policy.ID, policyID)
	}
	if policy.Name != "SDK Integration Test Policy" {
		t.Errorf("Name: got %q, want %q", policy.Name, "SDK Integration Test Policy")
	}
	if policy.Description != "Created by SDK integration test" {
		t.Errorf("Description: got %q, want %q", policy.Description, "Created by SDK integration test")
	}
	if policy.Category != dspm.CategoryOverexposed {
		t.Errorf("Category: got %q, want %q", policy.Category, dspm.CategoryOverexposed)
	}
	if policy.Severity != dspm.SeverityHigh {
		t.Errorf("Severity: got %q, want %q", policy.Severity, dspm.SeverityHigh)
	}
	if !policy.Enabled {
		t.Error("Enabled: got false, want true")
	}
	if policy.Predefined {
		t.Error("Predefined: got true, want false")
	}
	// Verify the filter structure matches the created policy.
	if policy.Filter == nil {
		t.Fatal("Filter is nil")
	}
	if policy.Filter.Op != dspm.LogicalAnd {
		t.Errorf("Filter.Op: got %q, want %q", policy.Filter.Op, dspm.LogicalAnd)
	}
	if len(policy.Filter.Filters) != 2 {
		t.Fatalf("Filter.Filters length: got %d, want 2", len(policy.Filter.Filters))
	}
	if policy.Filter.Filters[0].Config == nil {
		t.Fatal("Filter.Filters[0].Config is nil")
	}
	if policy.Filter.Filters[0].Config.Type != dspm.FilterTypeDocumentSensitivity {
		t.Errorf("Filter.Filters[0].Config.Type: got %q, want %q",
			policy.Filter.Filters[0].Config.Type, dspm.FilterTypeDocumentSensitivity)
	}
	if policy.Filter.Filters[1].GroupConfig == nil {
		t.Fatal("Filter.Filters[1].GroupConfig is nil")
	}
	if policy.Filter.Filters[1].GroupConfig.Op != dspm.LogicalOr {
		t.Errorf("Filter.Filters[1].GroupConfig.Op: got %q, want %q",
			policy.Filter.Filters[1].GroupConfig.Op, dspm.LogicalOr)
	}
	// Verify the threshold filter was unmarshaled.
	if policy.ThresholdFilter == nil {
		t.Fatal("ThresholdFilter is nil")
	}
	if policy.ThresholdFilter.Op != dspm.LogicalAnd {
		t.Errorf("ThresholdFilter.Op: got %q, want %q", policy.ThresholdFilter.Op, dspm.LogicalAnd)
	}
	if len(policy.ThresholdFilter.Filters) != 1 {
		t.Fatalf("ThresholdFilter.Filters length: got %d, want 1", len(policy.ThresholdFilter.Filters))
	}
	if policy.ThresholdFilter.Filters[0].Config == nil {
		t.Fatal("ThresholdFilter.Filters[0].Config is nil")
	}
	if policy.ThresholdFilter.Filters[0].Config.Type != dspm.FilterTypeDocumentExposure {
		t.Errorf("ThresholdFilter.Filters[0].Config.Type: got %q, want %q",
			policy.ThresholdFilter.Filters[0].Config.Type, dspm.FilterTypeDocumentExposure)
	}

	// Get policy by name.
	policy, err = dspmClient.PolicyByName(ctx, "SDK Integration Test Policy")
	if err != nil {
		t.Fatal(err)
	}
	if policy.ID != policyID {
		t.Errorf("PolicyByName ID: got %q, want %q", policy.ID, policyID)
	}

	// Verify PolicyByName returns ErrNotFound for non-existent name.
	_, err = dspmClient.PolicyByName(ctx, "Non-Existent Policy Name 12345")
	if err == nil || !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("expected graphql.ErrNotFound for non-existent name: %v", err)
	}

	// Verify PolicyByID returns an error for non-existent UUID.
	_, err = dspmClient.PolicyByID(ctx, uuid.Nil)
	if err == nil {
		t.Error("expected error for non-existent policy ID")
	}

	// List all policies and verify the created policy is present.
	policies, err := dspmClient.Policies(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range policies {
		if p.ID == policyID {
			found = true
			// Verify filter unmarshal through the list endpoint.
			if p.Filter == nil {
				t.Fatal("listed policy Filter is nil")
			}
			if len(p.Filter.Filters) != 2 {
				t.Errorf("listed policy Filter.Filters length: got %d, want 2",
					len(p.Filter.Filters))
			}
			if p.ThresholdFilter == nil {
				t.Fatal("listed policy ThresholdFilter is nil")
			}
			break
		}
	}
	if !found {
		t.Error("created policy not found in list")
	}

	// Update the policy with a different filter: single condition only.
	updatedName := "SDK Integration Test Policy Updated"
	updatedDesc := "Updated by SDK integration test"
	updatedSeverity := dspm.SeverityCritical
	updatedFilter := dspm.GroupConfig{
		Op: dspm.LogicalAnd,
		Filters: []dspm.Node{
			{
				Config: &dspm.Config{
					Type:         dspm.FilterTypeDocumentExposure,
					Values:       []string{exposureValue},
					Relationship: dspm.RelIs,
				},
			},
		},
	}
	err = dspmClient.UpdatePolicy(ctx, dspm.UpdateInput{
		ID:          policyID,
		Name:        &updatedName,
		Description: &updatedDesc,
		Severity:    &updatedSeverity,
		Filter:      &updatedFilter,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify the update.
	policy, err = dspmClient.PolicyByID(ctx, policyID)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Name != updatedName {
		t.Errorf("updated Name: got %q, want %q", policy.Name, updatedName)
	}
	if policy.Description != updatedDesc {
		t.Errorf("updated Description: got %q, want %q", policy.Description, updatedDesc)
	}
	if policy.Severity != dspm.SeverityCritical {
		t.Errorf("updated Severity: got %q, want %q", policy.Severity, dspm.SeverityCritical)
	}
	// Verify the updated filter was unmarshaled correctly.
	if policy.Filter == nil {
		t.Fatal("updated Filter is nil")
	}
	if policy.Filter.Op != dspm.LogicalAnd {
		t.Errorf("updated Filter.Op: got %q, want %q", policy.Filter.Op, dspm.LogicalAnd)
	}
	if len(policy.Filter.Filters) != 1 {
		t.Fatalf("updated Filter.Filters length: got %d, want 1", len(policy.Filter.Filters))
	}
	if policy.Filter.Filters[0].Config == nil {
		t.Fatal("updated Filter.Filters[0].Config is nil")
	}
	if policy.Filter.Filters[0].Config.Type != dspm.FilterTypeDocumentExposure {
		t.Errorf("updated Filter.Filters[0].Config.Type: got %q, want %q",
			policy.Filter.Filters[0].Config.Type, dspm.FilterTypeDocumentExposure)
	}
	// The API clears ThresholdFilter when it is omitted from the update.
	if policy.ThresholdFilter != nil {
		t.Error("ThresholdFilter: got non-nil after partial update, want nil")
	}

	// Disable the policy.
	disabled := false
	err = dspmClient.UpdatePolicy(ctx, dspm.UpdateInput{
		ID:      policyID,
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatal(err)
	}

	policy, err = dspmClient.PolicyByID(ctx, policyID)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Enabled {
		t.Error("Enabled after disable: got true, want false")
	}
}

func TestFilterTypesAndValues(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	dspmClient := Wrap(client)

	// Get filter types for object resources.
	types, err := dspmClient.FilterTypes(ctx, dspm.ResourceTypeObject)
	if err != nil {
		t.Fatal(err)
	}
	if len(types) == 0 {
		t.Fatal("expected at least one filter type for objects")
	}

	// Verify well-known filter types are present.
	found := false
	for _, ft := range types {
		if ft == dspm.FilterTypeDocumentSensitivity {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("%q not found in object filter types", dspm.FilterTypeDocumentSensitivity)
	}

	// Get possible values for a known filter type.
	meta, err := dspmClient.FilterValues(ctx, dspm.FilterTypeDocumentSensitivity, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(meta.Relationships) == 0 {
		t.Error("expected at least one relationship")
	}
	if len(meta.Values) == 0 {
		t.Error("expected at least one value for sensitivity filter")
	}

	// Verify that searchTerm is correctly passed through to the API by
	// searching for a known value label and checking that the result set
	// is filtered.
	searchLabel := meta.Values[0].Label
	filtered, err := dspmClient.FilterValues(ctx, dspm.FilterTypeDocumentSensitivity, searchLabel)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.Values) == 0 {
		t.Errorf("expected at least one value when searching for %q", searchLabel)
	}
	if len(filtered.Values) > len(meta.Values) {
		t.Errorf("filtered values (%d) should not exceed unfiltered values (%d)",
			len(filtered.Values), len(meta.Values))
	}
}
