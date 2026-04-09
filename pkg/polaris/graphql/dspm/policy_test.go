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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

func TestGroupConfigMarshalJSON(t *testing.T) {
	gc := GroupConfig{
		Op: LogicalAnd,
		Filters: []Node{
			{
				Config: &Config{
					Type:         FilterTypeDocumentDataType,
					Values:       []string{"password", "ssn"},
					Relationship: RelIs,
				},
			},
			{
				GroupConfig: &GroupConfig{
					Op: LogicalOr,
					Filters: []Node{
						{
							Config: &Config{
								Type:         FilterTypeSnappableName,
								Values:       []string{"server-1"},
								Relationship: RelContains,
							},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(gc)
	if err != nil {
		t.Fatalf("failed to marshal GroupConfig: %v", err)
	}

	var got GroupConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal GroupConfig: %v", err)
	}

	if got.Op != LogicalAnd {
		t.Errorf("Op: got %q, want %q", got.Op, LogicalAnd)
	}
	if len(got.Filters) != 2 {
		t.Fatalf("Filters length: got %d, want 2", len(got.Filters))
	}
	if got.Filters[0].Config == nil {
		t.Fatal("Filters[0].Config is nil")
	}
	if got.Filters[0].Config.Type != FilterTypeDocumentDataType {
		t.Errorf("Filters[0].Config.Type: got %q, want %q", got.Filters[0].Config.Type, FilterTypeDocumentDataType)
	}
	if len(got.Filters[0].Config.Values) != 2 {
		t.Fatalf("Filters[0].Config.Values length: got %d, want 2", len(got.Filters[0].Config.Values))
	}
	if got.Filters[1].GroupConfig == nil {
		t.Fatal("Filters[1].GroupConfig is nil")
	}
	if got.Filters[1].GroupConfig.Op != LogicalOr {
		t.Errorf("Filters[1].GroupConfig.Op: got %q, want %q", got.Filters[1].GroupConfig.Op, LogicalOr)
	}
}

func TestValidateFilterDepth(t *testing.T) {
	leaf := func() Node {
		return Node{Config: &Config{Type: FilterTypeDocumentDataType, Values: []string{"v"}, Relationship: RelIs}}
	}
	group := func(children ...Node) Node {
		return Node{GroupConfig: &GroupConfig{Op: LogicalAnd, Filters: children}}
	}

	// Depth 1: root group with leaves — OK.
	if err := validateFilterDepth(GroupConfig{Op: LogicalAnd, Filters: []Node{leaf()}}, 0); err != nil {
		t.Errorf("depth 1 should be valid: %v", err)
	}

	// Depth 2: root group containing a nested group with leaves — OK.
	if err := validateFilterDepth(GroupConfig{Op: LogicalAnd, Filters: []Node{group(leaf())}}, 0); err != nil {
		t.Errorf("depth 2 should be valid: %v", err)
	}

	// Depth 3: root -> group -> group -> leaf — should fail.
	if err := validateFilterDepth(GroupConfig{Op: LogicalAnd, Filters: []Node{group(group(leaf()))}}, 0); err == nil {
		t.Error("depth 3 should be invalid")
	}
}

func TestCreateInputMarshalJSON(t *testing.T) {
	input := CreateInput{
		Name:        "Test Policy",
		Description: "A test policy",
		Category:    CategoryMisplaced,
		Severity:    SeverityHigh,
		Filter: GroupConfig{
			Op: LogicalAnd,
			Filters: []Node{
				{
					Config: &Config{
						Type:         FilterTypeDocumentDataType,
						Values:       []string{"password"},
						Relationship: RelIs,
					},
				},
			},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal CreateInput: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	// Verify JSON field names match GraphQL input type.
	for _, key := range []string{"policyName", "description", "policyCategory", "policySeverity", "filter"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}

	// Verify thresholdFilter is omitted when nil.
	if _, ok := raw["thresholdFilter"]; ok {
		t.Error("thresholdFilter should be omitted when nil")
	}
}

func TestUpdateInputMarshalJSON(t *testing.T) {
	name := "Updated Name"
	enabled := false
	input := UpdateInput{
		ID:      uuid.MustParse("d4e5f6a7-b8c9-0123-4567-89abcdef0123"),
		Name:    &name,
		Enabled: &enabled,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal UpdateInput: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	// policyId is always present.
	if _, ok := raw["policyId"]; !ok {
		t.Error("missing policyId")
	}

	// Name and isEnabled are set.
	if _, ok := raw["policyName"]; !ok {
		t.Error("missing policyName")
	}
	if _, ok := raw["isEnabled"]; !ok {
		t.Error("missing isEnabled")
	}

	// Unset fields should be omitted.
	for _, key := range []string{"description", "policyCategory", "policySeverity", "filter", "thresholdFilter"} {
		if _, ok := raw[key]; ok {
			t.Errorf("field %q should be omitted when nil", key)
		}
	}
}

func TestCreatePolicy(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	policyID := "d4e5f6a7-b8c9-0123-4567-89abcdef0123"
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, `{"data":{"result":{"policyId":"%s"}}}`, policyID)
	}))
	defer srv.Close()

	id, err := CreatePolicy(ctx, graphql.NewTestClient(srv), CreateInput{
		Name:        "Test Policy",
		Description: "A test policy",
		Category:    CategoryMisplaced,
		Severity:    SeverityHigh,
		Filter: GroupConfig{
			Op: LogicalAnd,
			Filters: []Node{
				{Config: &Config{Type: FilterTypeDocumentDataType, Values: []string{"password"}, Relationship: RelIs}},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}
	if id != uuid.MustParse(policyID) {
		t.Errorf("PolicyID: got %q, want %q", id, policyID)
	}
}

func TestPolicyByID(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	policyID := "d4e5f6a7-b8c9-0123-4567-89abcdef0123"
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{
			"data": {
				"result": {
					"policy": {
						"policyId": "d4e5f6a7-b8c9-0123-4567-89abcdef0123",
						"name": "Test Policy",
						"description": "A test description",
						"policyCategory": "MISPLACED",
						"policySeverity": "HIGH",
						"isEnabled": true,
						"isPredefined": false,
						"filter": {
							"filterConfig": {
								"__typename": "FilterGroupConfig",
								"logicalOperator": "AND",
								"filtersList": [
									{
										"filterConfig": {
											"__typename": "FilterConfig",
											"type": "SECURITY_DOCUMENT_DATA_TYPE",
											"values": ["password"],
											"relationship": "IS"
										}
									}
								]
							}
						},
						"thresholdFilter": null,
						"createdAt": "2026-03-31T12:00:00Z",
						"updatedAt": "2026-03-31T12:00:00Z"
					}
				}
			}
		}`)
	}))
	defer srv.Close()

	policy, err := PolicyByID(ctx, graphql.NewTestClient(srv), uuid.MustParse(policyID))
	if err != nil {
		t.Fatalf("PolicyByID failed: %v", err)
	}
	if policy.ID != uuid.MustParse(policyID) {
		t.Errorf("ID: got %q, want %q", policy.ID, policyID)
	}
	if policy.Name != "Test Policy" {
		t.Errorf("Name: got %q, want %q", policy.Name, "Test Policy")
	}
	if policy.Description != "A test description" {
		t.Errorf("Description: got %q, want %q", policy.Description, "A test description")
	}
	if policy.Category != CategoryMisplaced {
		t.Errorf("Category: got %q, want %q", policy.Category, CategoryMisplaced)
	}
	if policy.Severity != SeverityHigh {
		t.Errorf("Severity: got %q, want %q", policy.Severity, SeverityHigh)
	}
	if !policy.Enabled {
		t.Error("Enabled: got false, want true")
	}
	if policy.Predefined {
		t.Error("Predefined: got true, want false")
	}
	if policy.Filter == nil {
		t.Fatal("Filter is nil")
	}
	if policy.Filter.Op != LogicalAnd {
		t.Errorf("Filter.Op: got %q, want %q", policy.Filter.Op, LogicalAnd)
	}
	if len(policy.Filter.Filters) != 1 {
		t.Fatalf("Filter.Filters length: got %d, want 1", len(policy.Filter.Filters))
	}
	if policy.Filter.Filters[0].Config == nil {
		t.Fatal("Filter.Filters[0].Config is nil")
	}
	if policy.Filter.Filters[0].Config.Type != FilterTypeDocumentDataType {
		t.Errorf("Filter.Filters[0].Config.Type: got %q, want %q",
			policy.Filter.Filters[0].Config.Type, FilterTypeDocumentDataType)
	}
	if len(policy.Filter.Filters[0].Config.Values) != 1 || policy.Filter.Filters[0].Config.Values[0] != "password" {
		t.Errorf("Filter.Filters[0].Config.Values: got %v, want [password]",
			policy.Filter.Filters[0].Config.Values)
	}
	if policy.Filter.Filters[0].Config.Relationship != RelIs {
		t.Errorf("Filter.Filters[0].Config.Relationship: got %q, want %q",
			policy.Filter.Filters[0].Config.Relationship, RelIs)
	}
	if policy.ThresholdFilter != nil {
		t.Error("ThresholdFilter: got non-nil, want nil")
	}
}

func TestPolicyUnmarshalJSON(t *testing.T) {
	// Nested groups (depth 2).
	nestedJSON := `{
		"policyId": "d4e5f6a7-b8c9-0123-4567-89abcdef0123",
		"name": "Test",
		"description": "",
		"policyCategory": "MISPLACED",
		"policySeverity": "HIGH",
		"isEnabled": true,
		"isPredefined": false,
		"filter": {
			"filterConfig": {
				"__typename": "FilterGroupConfig",
				"logicalOperator": "AND",
				"filtersList": [
					{
						"filterConfig": {
							"__typename": "FilterConfig",
							"type": "SECURITY_DOCUMENT_DATA_TYPE",
							"values": ["password"],
							"relationship": "IS"
						}
					},
					{
						"filterConfig": {
							"__typename": "FilterGroupConfig",
							"logicalOperator": "OR",
							"filtersList": [
								{
									"filterConfig": {
										"__typename": "FilterConfig",
										"type": "SECURITY_SNAPPABLE_NAME",
										"values": ["server-1"],
										"relationship": "CONTAINS"
									}
								}
							]
						}
					}
				]
			}
		},
		"thresholdFilter": null,
		"createdAt": "2026-03-31T12:00:00Z",
		"updatedAt": "2026-03-31T12:00:00Z"
	}`

	var policy Policy
	if err := json.Unmarshal([]byte(nestedJSON), &policy); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if policy.Filter == nil {
		t.Fatal("Filter is nil")
	}
	if policy.Filter.Op != LogicalAnd {
		t.Errorf("Filter.Op: got %q, want %q", policy.Filter.Op, LogicalAnd)
	}
	if len(policy.Filter.Filters) != 2 {
		t.Fatalf("Filter.Filters length: got %d, want 2", len(policy.Filter.Filters))
	}
	if policy.Filter.Filters[0].Config == nil {
		t.Fatal("Filters[0].Config is nil")
	}
	if policy.Filter.Filters[0].Config.Type != FilterTypeDocumentDataType {
		t.Errorf("Filters[0].Config.Type: got %q, want %q",
			policy.Filter.Filters[0].Config.Type, FilterTypeDocumentDataType)
	}
	if policy.Filter.Filters[1].GroupConfig == nil {
		t.Fatal("Filters[1].GroupConfig is nil")
	}
	if policy.Filter.Filters[1].GroupConfig.Op != LogicalOr {
		t.Errorf("Filters[1].GroupConfig.Op: got %q, want %q",
			policy.Filter.Filters[1].GroupConfig.Op, LogicalOr)
	}
	if len(policy.Filter.Filters[1].GroupConfig.Filters) != 1 {
		t.Fatalf("Filters[1].GroupConfig.Filters length: got %d, want 1",
			len(policy.Filter.Filters[1].GroupConfig.Filters))
	}
	if policy.ThresholdFilter != nil {
		t.Error("ThresholdFilter: got non-nil, want nil")
	}

	// Null filter.
	nullJSON := `{
		"policyId": "d4e5f6a7-b8c9-0123-4567-89abcdef0123",
		"name": "Test",
		"description": "",
		"policyCategory": "MISPLACED",
		"policySeverity": "HIGH",
		"isEnabled": true,
		"isPredefined": false,
		"filter": null,
		"thresholdFilter": null,
		"createdAt": "2026-03-31T12:00:00Z",
		"updatedAt": "2026-03-31T12:00:00Z"
	}`

	var nullPolicy Policy
	if err := json.Unmarshal([]byte(nullJSON), &nullPolicy); err != nil {
		t.Fatalf("Unmarshal null filter failed: %v", err)
	}
	if nullPolicy.Filter != nil {
		t.Error("null filter: got non-nil, want nil")
	}

	// Unknown __typename.
	unknownJSON := `{
		"policyId": "d4e5f6a7-b8c9-0123-4567-89abcdef0123",
		"name": "Test",
		"description": "",
		"policyCategory": "MISPLACED",
		"policySeverity": "HIGH",
		"isEnabled": true,
		"isPredefined": false,
		"filter": {
			"filterConfig": {
				"__typename": "UnknownType"
			}
		},
		"thresholdFilter": null,
		"createdAt": "2026-03-31T12:00:00Z",
		"updatedAt": "2026-03-31T12:00:00Z"
	}`

	var unknownPolicy Policy
	if err := json.Unmarshal([]byte(unknownJSON), &unknownPolicy); err == nil {
		t.Error("expected error for unknown __typename")
	}

	// Lone FilterConfig at top level — exercises the wrapping fallback in
	// parsePolicyFilter.
	loneConfigJSON := `{
		"policyId": "d4e5f6a7-b8c9-0123-4567-89abcdef0123",
		"name": "Test",
		"description": "",
		"policyCategory": "MISPLACED",
		"policySeverity": "HIGH",
		"isEnabled": true,
		"isPredefined": false,
		"filter": {
			"filterConfig": {
				"__typename": "FilterConfig",
				"type": "SECURITY_DOCUMENT_DATA_TYPE",
				"values": ["password"],
				"relationship": "IS"
			}
		},
		"thresholdFilter": null,
		"createdAt": "2026-03-31T12:00:00Z",
		"updatedAt": "2026-03-31T12:00:00Z"
	}`

	var lonePolicy Policy
	if err := json.Unmarshal([]byte(loneConfigJSON), &lonePolicy); err != nil {
		t.Fatalf("Unmarshal lone FilterConfig failed: %v", err)
	}
	if lonePolicy.Filter == nil {
		t.Fatal("lone FilterConfig: Filter is nil")
	}
	if lonePolicy.Filter.Op != LogicalAnd {
		t.Errorf("lone FilterConfig: Op: got %q, want %q", lonePolicy.Filter.Op, LogicalAnd)
	}
	if len(lonePolicy.Filter.Filters) != 1 {
		t.Fatalf("lone FilterConfig: Filters length: got %d, want 1", len(lonePolicy.Filter.Filters))
	}
	if lonePolicy.Filter.Filters[0].Config == nil {
		t.Fatal("lone FilterConfig: Filters[0].Config is nil")
	}
	if lonePolicy.Filter.Filters[0].Config.Type != FilterTypeDocumentDataType {
		t.Errorf("lone FilterConfig: Filters[0].Config.Type: got %q, want %q",
			lonePolicy.Filter.Filters[0].Config.Type, FilterTypeDocumentDataType)
	}

	// Null filterConfig inside a non-null filter envelope.
	nullConfigJSON := `{
		"policyId": "d4e5f6a7-b8c9-0123-4567-89abcdef0123",
		"name": "Test",
		"description": "",
		"policyCategory": "MISPLACED",
		"policySeverity": "HIGH",
		"isEnabled": true,
		"isPredefined": false,
		"filter": {"filterConfig": null},
		"thresholdFilter": null,
		"createdAt": "2026-03-31T12:00:00Z",
		"updatedAt": "2026-03-31T12:00:00Z"
	}`

	var nullConfigPolicy Policy
	if err := json.Unmarshal([]byte(nullConfigJSON), &nullConfigPolicy); err != nil {
		t.Fatalf("Unmarshal null filterConfig failed: %v", err)
	}
	if nullConfigPolicy.Filter != nil {
		t.Error("null filterConfig: got non-nil, want nil")
	}

	// Both Filter and ThresholdFilter non-null.
	bothJSON := `{
		"policyId": "d4e5f6a7-b8c9-0123-4567-89abcdef0123",
		"name": "Test",
		"description": "",
		"policyCategory": "MISPLACED",
		"policySeverity": "HIGH",
		"isEnabled": true,
		"isPredefined": false,
		"filter": {
			"filterConfig": {
				"__typename": "FilterGroupConfig",
				"logicalOperator": "AND",
				"filtersList": [
					{
						"filterConfig": {
							"__typename": "FilterConfig",
							"type": "SECURITY_DOCUMENT_DATA_TYPE",
							"values": ["password"],
							"relationship": "IS"
						}
					}
				]
			}
		},
		"thresholdFilter": {
			"filterConfig": {
				"__typename": "FilterGroupConfig",
				"logicalOperator": "OR",
				"filtersList": [
					{
						"filterConfig": {
							"__typename": "FilterConfig",
							"type": "SECURITY_SNAPPABLE_NAME",
							"values": ["server-1"],
							"relationship": "CONTAINS"
						}
					}
				]
			}
		},
		"createdAt": "2026-03-31T12:00:00Z",
		"updatedAt": "2026-03-31T12:00:00Z"
	}`

	var bothPolicy Policy
	if err := json.Unmarshal([]byte(bothJSON), &bothPolicy); err != nil {
		t.Fatalf("Unmarshal both filters failed: %v", err)
	}
	if bothPolicy.Filter == nil {
		t.Fatal("both: Filter is nil")
	}
	if bothPolicy.Filter.Op != LogicalAnd {
		t.Errorf("both: Filter.Op: got %q, want %q", bothPolicy.Filter.Op, LogicalAnd)
	}
	if bothPolicy.ThresholdFilter == nil {
		t.Fatal("both: ThresholdFilter is nil")
	}
	if bothPolicy.ThresholdFilter.Op != LogicalOr {
		t.Errorf("both: ThresholdFilter.Op: got %q, want %q", bothPolicy.ThresholdFilter.Op, LogicalOr)
	}
	if len(bothPolicy.ThresholdFilter.Filters) != 1 {
		t.Fatalf("both: ThresholdFilter.Filters length: got %d, want 1",
			len(bothPolicy.ThresholdFilter.Filters))
	}
	if bothPolicy.ThresholdFilter.Filters[0].Config == nil {
		t.Fatal("both: ThresholdFilter.Filters[0].Config is nil")
	}
	if bothPolicy.ThresholdFilter.Filters[0].Config.Type != FilterTypeSnappableName {
		t.Errorf("both: ThresholdFilter.Filters[0].Config.Type: got %q, want %q",
			bothPolicy.ThresholdFilter.Filters[0].Config.Type, FilterTypeSnappableName)
	}

	// Malformed JSON inside filter.
	malformedJSON := `{
		"policyId": "d4e5f6a7-b8c9-0123-4567-89abcdef0123",
		"name": "Test",
		"description": "",
		"policyCategory": "MISPLACED",
		"policySeverity": "HIGH",
		"isEnabled": true,
		"isPredefined": false,
		"filter": {"filterConfig": "not-an-object"},
		"thresholdFilter": null,
		"createdAt": "2026-03-31T12:00:00Z",
		"updatedAt": "2026-03-31T12:00:00Z"
	}`

	var malformedPolicy Policy
	if err := json.Unmarshal([]byte(malformedJSON), &malformedPolicy); err == nil {
		t.Error("expected error for malformed filterConfig")
	}
}

func TestPolicies(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{
			"data": {
				"result": [
					{
						"policy": {
							"policyId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
							"name": "Test Policy 1",
							"description": "First test policy",
							"policyCategory": "MISPLACED",
							"policySeverity": "HIGH",
							"isEnabled": true,
							"isPredefined": false,
							"filter": {
								"filterConfig": {
									"__typename": "FilterGroupConfig",
									"logicalOperator": "AND",
									"filtersList": [
										{
											"filterConfig": {
												"__typename": "FilterConfig",
												"type": "SECURITY_DOCUMENT_DATA_TYPE",
												"values": ["password"],
												"relationship": "IS"
											}
										}
									]
								}
							},
							"thresholdFilter": null,
							"createdAt": "2026-03-31T12:00:00Z",
							"updatedAt": "2026-03-31T12:00:00Z"
						}
					},
					{
						"policy": {
							"policyId": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
							"name": "Test Policy 2",
							"description": "Second test policy",
							"policyCategory": "OVEREXPOSED",
							"policySeverity": "CRITICAL",
							"isEnabled": false,
							"isPredefined": true,
							"filter": null,
							"thresholdFilter": null,
							"createdAt": "2026-03-30T12:00:00Z",
							"updatedAt": "2026-03-30T12:00:00Z"
						}
					}
				]
			}
		}`)
	}))
	defer srv.Close()

	policies, err := Policies(ctx, graphql.NewTestClient(srv))
	if err != nil {
		t.Fatalf("Policies failed: %v", err)
	}
	if len(policies) != 2 {
		t.Fatalf("got %d policies, want 2", len(policies))
	}
	if policies[0].Name != "Test Policy 1" {
		t.Errorf("policies[0].Name: got %q, want %q", policies[0].Name, "Test Policy 1")
	}
	if policies[0].Category != CategoryMisplaced {
		t.Errorf("policies[0].Category: got %q, want %q", policies[0].Category, CategoryMisplaced)
	}
	if policies[0].Filter == nil {
		t.Fatal("policies[0].Filter is nil")
	}
	if policies[0].Filter.Op != LogicalAnd {
		t.Errorf("policies[0].Filter.Op: got %q, want %q", policies[0].Filter.Op, LogicalAnd)
	}
	if len(policies[0].Filter.Filters) != 1 {
		t.Fatalf("policies[0].Filter.Filters length: got %d, want 1", len(policies[0].Filter.Filters))
	}
	if policies[1].Filter != nil {
		t.Error("policies[1].Filter: got non-nil, want nil")
	}
	if policies[1].Name != "Test Policy 2" {
		t.Errorf("policies[1].Name: got %q, want %q", policies[1].Name, "Test Policy 2")
	}
	if policies[1].Category != CategoryOverexposed {
		t.Errorf("policies[1].Category: got %q, want %q", policies[1].Category, CategoryOverexposed)
	}
	if !policies[1].Predefined {
		t.Error("policies[1].Predefined: got false, want true")
	}
}

func TestUpdatePolicy(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{"data":{"result":null}}`)
	}))
	defer srv.Close()

	name := "Updated Policy"
	err := UpdatePolicy(ctx, graphql.NewTestClient(srv), UpdateInput{
		ID:   uuid.MustParse("d4e5f6a7-b8c9-0123-4567-89abcdef0123"),
		Name: &name,
	})
	if err != nil {
		t.Fatalf("UpdatePolicy failed: %v", err)
	}
}

func TestDeletePolicy(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{"data":{"result":null}}`)
	}))
	defer srv.Close()

	err := DeletePolicy(ctx, graphql.NewTestClient(srv), uuid.MustParse("d4e5f6a7-b8c9-0123-4567-89abcdef0123"))
	if err != nil {
		t.Fatalf("DeletePolicy failed: %v", err)
	}
}
