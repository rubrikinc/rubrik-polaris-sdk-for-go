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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

func TestFilterTypes(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{
			"data": {
				"result": [
					"SECURITY_SNAPPABLE_NAME",
					"SECURITY_SNAPPABLE_TAG",
					"SECURITY_SNAPPABLE_BACKUP",
					"SECURITY_SNAPPABLE_ENCRYPTION",
					"SECURITY_DOCUMENT_DATA_TYPE",
					"SECURITY_DOCUMENT_SENSITIVITY"
				]
			}
		}`)
	}))
	defer srv.Close()

	types, err := FilterTypes(ctx, graphql.NewTestClient(srv), ResourceTypeObject)
	if err != nil {
		t.Fatalf("FilterTypes failed: %v", err)
	}
	if len(types) != 6 {
		t.Fatalf("got %d filter types, want 6", len(types))
	}
	if types[0] != FilterTypeSnappableName {
		t.Errorf("types[0]: got %q, want %q", types[0], FilterTypeSnappableName)
	}
	if types[4] != FilterTypeDocumentDataType {
		t.Errorf("types[4]: got %q, want %q", types[4], FilterTypeDocumentDataType)
	}
}

func TestFilterValues(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{
			"data": {
				"result": {
					"possibleRelationships": ["IS", "IS_NOT"],
					"possibleValues": {
						"__typename": "FilterValues",
						"filterValues": [
							{"id": "HIGH", "label": "High"},
							{"id": "MEDIUM", "label": "Medium"},
							{"id": "LOW", "label": "Low"},
							{"id": "NON_SENSITIVE", "label": "Non-Sensitive"}
						]
					}
				}
			}
		}`)
	}))
	defer srv.Close()

	meta, err := FilterValues(ctx, graphql.NewTestClient(srv), FilterTypeDocumentSensitivity, "")
	if err != nil {
		t.Fatalf("FilterValues failed: %v", err)
	}
	if len(meta.Relationships) != 2 {
		t.Fatalf("got %d relationships, want 2", len(meta.Relationships))
	}
	if meta.Relationships[0] != RelIs {
		t.Errorf("relationships[0]: got %q, want %q", meta.Relationships[0], RelIs)
	}
	if meta.Relationships[1] != RelIsNot {
		t.Errorf("relationships[1]: got %q, want %q", meta.Relationships[1], RelIsNot)
	}
	if len(meta.Values) != 4 {
		t.Fatalf("got %d values, want 4", len(meta.Values))
	}
	if meta.Values[0].ID != "HIGH" {
		t.Errorf("values[0].ID: got %q, want %q", meta.Values[0].ID, "HIGH")
	}
	if meta.Values[0].Label != "High" {
		t.Errorf("values[0].Label: got %q, want %q", meta.Values[0].Label, "High")
	}
}

func TestFilterTreeValues(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{
			"data": {
				"result": {
					"possibleRelationships": ["IS", "IS_NOT"],
					"possibleValues": {
						"__typename": "FilterTreeValues",
						"filterValues": [
							{
								"value": {"id": "PARENT_1", "label": "Parent One"},
								"children": [
									{"value": {"id": "CHILD_1A", "label": "Child 1A"}},
									{"value": {"id": "CHILD_1B", "label": "Child 1B"}}
								]
							},
							{
								"value": {"id": "PARENT_2", "label": "Parent Two"},
								"children": []
							}
						]
					}
				}
			}
		}`)
	}))
	defer srv.Close()

	meta, err := FilterValues(ctx, graphql.NewTestClient(srv), FilterTypeSnappableTag, "")
	if err != nil {
		t.Fatalf("FilterValues failed: %v", err)
	}
	if len(meta.Relationships) != 2 {
		t.Fatalf("got %d relationships, want 2", len(meta.Relationships))
	}
	if len(meta.TreeValues) != 2 {
		t.Fatalf("got %d tree values, want 2", len(meta.TreeValues))
	}
	if meta.TreeValues[0].Value.ID != "PARENT_1" {
		t.Errorf("tree[0].Value.ID: got %q, want %q", meta.TreeValues[0].Value.ID, "PARENT_1")
	}
	if meta.TreeValues[0].Value.Label != "Parent One" {
		t.Errorf("tree[0].Value.Label: got %q, want %q", meta.TreeValues[0].Value.Label, "Parent One")
	}
	if len(meta.TreeValues[0].Children) != 2 {
		t.Fatalf("tree[0] children: got %d, want 2", len(meta.TreeValues[0].Children))
	}
	if meta.TreeValues[0].Children[0].Value.ID != "CHILD_1A" {
		t.Errorf("tree[0].Children[0].Value.ID: got %q, want %q", meta.TreeValues[0].Children[0].Value.ID, "CHILD_1A")
	}
	if meta.TreeValues[0].Children[1].Value.ID != "CHILD_1B" {
		t.Errorf("tree[0].Children[1].Value.ID: got %q, want %q", meta.TreeValues[0].Children[1].Value.ID, "CHILD_1B")
	}
	if len(meta.TreeValues[1].Children) != 0 {
		t.Errorf("tree[1] children: got %d, want 0", len(meta.TreeValues[1].Children))
	}
	if len(meta.Values) != 0 {
		t.Errorf("flat Values should be empty, got %d", len(meta.Values))
	}
	if len(meta.ProviderValues) != 0 {
		t.Errorf("ProviderValues should be empty, got %d", len(meta.ProviderValues))
	}
}

func TestFilterValuesWithProvider(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{
			"data": {
				"result": {
					"possibleRelationships": ["IS", "IS_NOT"],
					"possibleValues": {
						"__typename": "FilterValuesWithProvider",
						"filterValuesWithProvider": [
							{
								"filterValue": {"id": "ACCOUNT_1", "label": "Account One"},
								"provider": "AWS"
							},
							{
								"filterValue": {"id": "ACCOUNT_2", "label": "Account Two"},
								"provider": "AZURE"
							}
						]
					}
				}
			}
		}`)
	}))
	defer srv.Close()

	meta, err := FilterValues(ctx, graphql.NewTestClient(srv), FilterTypeSnappableCloudAccount, "")
	if err != nil {
		t.Fatalf("FilterValues failed: %v", err)
	}
	if len(meta.Relationships) != 2 {
		t.Fatalf("got %d relationships, want 2", len(meta.Relationships))
	}
	if len(meta.ProviderValues) != 2 {
		t.Fatalf("got %d provider values, want 2", len(meta.ProviderValues))
	}
	if meta.ProviderValues[0].FilterValue.ID != "ACCOUNT_1" {
		t.Errorf("provider[0].Value.ID: got %q, want %q", meta.ProviderValues[0].FilterValue.ID, "ACCOUNT_1")
	}
	if meta.ProviderValues[0].FilterValue.Label != "Account One" {
		t.Errorf("provider[0].Value.Label: got %q, want %q", meta.ProviderValues[0].FilterValue.Label, "Account One")
	}
	if meta.ProviderValues[0].Provider != "AWS" {
		t.Errorf("provider[0].Provider: got %q, want %q", meta.ProviderValues[0].Provider, "AWS")
	}
	if meta.ProviderValues[1].FilterValue.ID != "ACCOUNT_2" {
		t.Errorf("provider[1].Value.ID: got %q, want %q", meta.ProviderValues[1].FilterValue.ID, "ACCOUNT_2")
	}
	if meta.ProviderValues[1].Provider != "AZURE" {
		t.Errorf("provider[1].Provider: got %q, want %q", meta.ProviderValues[1].Provider, "AZURE")
	}
	if len(meta.Values) != 0 {
		t.Errorf("flat Values should be empty, got %d", len(meta.Values))
	}
	if len(meta.TreeValues) != 0 {
		t.Errorf("TreeValues should be empty, got %d", len(meta.TreeValues))
	}
}
