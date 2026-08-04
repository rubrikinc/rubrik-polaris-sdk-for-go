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

package hierarchy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	azureregions "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// TestObjectsByName verifies that ObjectsByName sets the NAME_EXACT_MATCH filter
// on the inventory query and unmarshals the returned nodes.
func TestObjectsByName(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	const response = `{
		"data": {
			"result": {
				"descendantConnection": {
					"count": 1,
					"nodes": [
						{
							"id": "11111111-1111-1111-1111-111111111111",
							"name": "example-vm",
							"objectType": "AzureNativeVm"
						}
					],
					"pageInfo": {"endCursor": "", "hasNextPage": false}
				}
			}
		}
	}`

	var gotFilter []Filter
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var decoded struct {
			Variables struct {
				Filter []Filter `json:"filter"`
			} `json:"variables"`
		}
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		gotFilter = decoded.Variables.Filter
		io.WriteString(w, response)
	}))
	defer srv.Close()

	objects, err := ObjectsByName[AzureNativeVirtualMachine](ctx, Wrap(graphql.NewTestClient(srv)), "example-vm", WorkloadAllSubHierarchyType)
	if err != nil {
		t.Fatalf("ObjectsByName failed: %v", err)
	}

	if len(gotFilter) != 1 || gotFilter[0].Field != "NAME_EXACT_MATCH" ||
		len(gotFilter[0].Texts) != 1 || gotFilter[0].Texts[0] != "example-vm" {
		t.Errorf("filter: got %+v, want NAME_EXACT_MATCH=[example-vm]", gotFilter)
	}

	if len(objects) != 1 {
		t.Fatalf("got %d objects, want 1", len(objects))
	}
	if objects[0].ID != uuid.MustParse("11111111-1111-1111-1111-111111111111") {
		t.Errorf("id: got %q", objects[0].ID)
	}
	if objects[0].Name != "example-vm" {
		t.Errorf("name: got %q", objects[0].Name)
	}
}

// TestAzureSQLManagedInstanceServer verifies the type's inventory type filter
// and that an API response unmarshals into the type, including the parent
// subscription details.
func TestAzureSQLManagedInstanceServer(t *testing.T) {
	// InventoryObject is a constraint interface, so it can't be used as a
	// variable type; call typeFilter directly on the value instead.
	if typeFilter := (AzureSQLManagedInstanceServer{}).typeFilter(); typeFilter != "AzureSqlManagedInstanceServer" {
		t.Fatalf("wrong type filter: %s", typeFilter)
	}

	const res = `{
		"id": "11111111-1111-1111-1111-111111111111",
		"name": "example-sql-mi",
		"objectType": "AzureSqlManagedInstanceServer",
		"azureResourceGroupDetails": {
			"azureSubscriptionDetails": {
				"id": "22222222-2222-2222-2222-222222222222",
				"name": "example-subscription",
				"accountConnectionId": "33333333-3333-3333-3333-333333333333",
				"tenantId": "44444444-4444-4444-4444-444444444444",
				"cloudType": "AZUREPUBLICCLOUD",
				"status": "REFRESHED",
				"regionSpecs": [{"region": "EASTUS2"}]
			}
		}
	}`

	var obj AzureSQLManagedInstanceServer
	if err := json.Unmarshal([]byte(res), &obj); err != nil {
		t.Fatal(err)
	}

	if obj.Name != "example-sql-mi" {
		t.Errorf("name: got %q", obj.Name)
	}
	if obj.ObjectType != "AzureSqlManagedInstanceServer" {
		t.Errorf("objectType: got %q", obj.ObjectType)
	}

	sub := obj.ResourceGroup.Subscription
	if sub.CloudAccountID != uuid.MustParse("33333333-3333-3333-3333-333333333333") {
		t.Errorf("subscription cloud account ID: got %q", sub.CloudAccountID)
	}
	if sub.Cloud != "AZUREPUBLICCLOUD" {
		t.Errorf("subscription cloud: got %q, want %q", sub.Cloud, "AZUREPUBLICCLOUD")
	}
	if len(sub.RegionSpecs) != 1 ||
		sub.RegionSpecs[0].Region.Region != azureregions.RegionFromNativeRegionEnum("EASTUS2") {
		t.Errorf("region specs: got %+v", sub.RegionSpecs)
	}
}
