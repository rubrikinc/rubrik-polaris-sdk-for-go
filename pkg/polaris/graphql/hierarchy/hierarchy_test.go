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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	awsregions "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
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

// TestAzurePostgresFlexibleServer verifies the type's inventory type filter and
// that an API response unmarshals into the type, including the parent
// subscription details.
func TestAzurePostgresFlexibleServer(t *testing.T) {
	// InventoryObject is a constraint interface, so it can't be used as a
	// variable type; call typeFilter directly on the value instead. Note that
	// RSC spells this HierarchyObjectTypeEnum value in SCREAMING_SNAKE_CASE,
	// unlike the CamelCase AzureSqlManagedInstanceServer value.
	if typeFilter := (AzurePostgresFlexibleServer{}).typeFilter(); typeFilter != "AZURE_POSTGRES_FLEXIBLE_SERVER" {
		t.Fatalf("wrong type filter: %s", typeFilter)
	}

	const res = `{
		"id": "11111111-1111-1111-1111-111111111111",
		"name": "example-flex-server",
		"objectType": "AZURE_POSTGRES_FLEXIBLE_SERVER",
		"azureResourceGroupDetails": {
			"azureSubscriptionDetails": {
				"id": "22222222-2222-2222-2222-222222222222",
				"name": "example-subscription",
				"accountConnectionId": "33333333-3333-3333-3333-333333333333",
				"tenantId": "example.onmicrosoft.com",
				"cloudType": "AZUREPUBLICCLOUD",
				"status": "REFRESHED",
				"regionSpecs": [{"region": "EAST_US2"}]
			}
		}
	}`

	var obj AzurePostgresFlexibleServer
	if err := json.Unmarshal([]byte(res), &obj); err != nil {
		t.Fatal(err)
	}

	if obj.Name != "example-flex-server" {
		t.Errorf("name: got %q", obj.Name)
	}
	if obj.ObjectType != "AZURE_POSTGRES_FLEXIBLE_SERVER" {
		t.Errorf("objectType: got %q", obj.ObjectType)
	}

	// The subscription details drive the client-side subscription filtering
	// callers do, since the inventory query has no subscription filter.
	sub := obj.ResourceGroup.Subscription
	if sub.CloudAccountID != uuid.MustParse("33333333-3333-3333-3333-333333333333") {
		t.Errorf("subscription cloud account ID: got %q", sub.CloudAccountID)
	}
	// RSC returns the tenant domain, not a UUID, in the tenantId field.
	if sub.TenantDomain != "example.onmicrosoft.com" {
		t.Errorf("subscription tenant domain: got %q", sub.TenantDomain)
	}
	if sub.Cloud != "AZUREPUBLICCLOUD" {
		t.Errorf("subscription cloud: got %q, want %q", sub.Cloud, "AZUREPUBLICCLOUD")
	}
	// Compare against the region constant rather than round-tripping the same
	// string through RegionFromNativeRegionEnum, which would pass vacuously if
	// both sides resolved to RegionUnknown.
	if len(sub.RegionSpecs) != 1 || sub.RegionSpecs[0].Region.Region != azureregions.RegionEastUS2 {
		t.Errorf("region specs: got %+v, want [EAST_US2]", sub.RegionSpecs)
	}
}

// TestAzureSQLManagedInstanceServer verifies the type's inventory type filter
// and that an API response unmarshals into the type, including the supported
// authentication mechanisms and the parent subscription details.
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
		"authType": "SQL_AUTH_AND_AAD",
		"azureResourceGroupDetails": {
			"azureSubscriptionDetails": {
				"id": "22222222-2222-2222-2222-222222222222",
				"name": "example-subscription",
				"accountConnectionId": "33333333-3333-3333-3333-333333333333",
				"tenantId": "example.onmicrosoft.com",
				"cloudType": "AZUREPUBLICCLOUD",
				"status": "REFRESHED",
				"regionSpecs": [{"region": "EAST_US2"}]
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
	if obj.AuthType != "SQL_AUTH_AND_AAD" {
		t.Errorf("authType: got %q, want %q", obj.AuthType, "SQL_AUTH_AND_AAD")
	}

	sub := obj.ResourceGroup.Subscription
	if sub.CloudAccountID != uuid.MustParse("33333333-3333-3333-3333-333333333333") {
		t.Errorf("subscription cloud account ID: got %q", sub.CloudAccountID)
	}
	// RSC returns the tenant domain, not a UUID, in the tenantId field.
	if sub.TenantDomain != "example.onmicrosoft.com" {
		t.Errorf("subscription tenant domain: got %q", sub.TenantDomain)
	}
	if sub.Cloud != "AZUREPUBLICCLOUD" {
		t.Errorf("subscription cloud: got %q, want %q", sub.Cloud, "AZUREPUBLICCLOUD")
	}
	// Compare against the region constant rather than round-tripping the same
	// string through RegionFromNativeRegionEnum, which would pass vacuously if
	// both sides resolved to RegionUnknown.
	if len(sub.RegionSpecs) != 1 || sub.RegionSpecs[0].Region.Region != azureregions.RegionEastUS2 {
		t.Errorf("region specs: got %+v, want [EAST_US2]", sub.RegionSpecs)
	}
}

// notFoundResponse is the shape RSC answers with for a hierarchy object which
// does not exist. Note that the HTTP status is 200 and the 404 appears only in
// the GraphQL error extensions.
const notFoundResponse = `{
	"data": null,
	"errors": [
		{
			"message": "NOT_FOUND: Unable to find managed object for ID or you are not authorized to access it",
			"extensions": {
				"code": 404,
				"trace": {"traceId": "kRKX9KBaV6v2HFaefdkynQ=="}
			}
		}
	]
}`

// TestObjectByIDAndWorkloadNotFound verifies that a missing hierarchy object is
// reported as graphql.ErrNotFound, so callers can match on the sentinel, and
// that the underlying GQLError is still reachable so the RSC trace ID is not
// lost.
func TestObjectByIDAndWorkloadNotFound(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, notFoundResponse)
	}))
	defer srv.Close()

	_, err := ObjectByIDAndWorkload[Object](ctx, graphql.NewTestClient(srv),
		uuid.MustParse("11111111-1111-1111-1111-111111111111"), WorkloadAllSubHierarchyType)
	if err == nil {
		t.Fatal("ObjectByIDAndWorkload succeeded, want an error")
	}
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("got %v, want it to match graphql.ErrNotFound", err)
	}

	var gqlErr graphql.GQLError
	if !errors.As(err, &gqlErr) {
		t.Fatalf("got %v, want the GQLError to remain in the chain", err)
	}
	if gqlErr.Code() != http.StatusNotFound {
		t.Errorf("code: got %d, want %d", gqlErr.Code(), http.StatusNotFound)
	}
}

// TestObjectByIDAndWorkloadOtherErrorIsNotNotFound verifies that an error which
// is not a 404 is not reported as graphql.ErrNotFound. Without this, a
// permission or server failure would be mistaken for a deleted object.
func TestObjectByIDAndWorkloadOtherErrorIsNotNotFound(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	const forbiddenResponse = `{
		"data": null,
		"errors": [
			{
				"message": "Objects are not authorized.",
				"extensions": {"code": 403, "trace": {"traceId": "mGREPE/AMVXl+dlY9A5+pg=="}}
			}
		]
	}`

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, forbiddenResponse)
	}))
	defer srv.Close()

	_, err := ObjectByIDAndWorkload[Object](ctx, graphql.NewTestClient(srv),
		uuid.MustParse("11111111-1111-1111-1111-111111111111"), WorkloadAllSubHierarchyType)
	if err == nil {
		t.Fatal("ObjectByIDAndWorkload succeeded, want an error")
	}
	if errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("got %v, want it not to match graphql.ErrNotFound", err)
	}
}

// TestIcebergObjectsByType verifies that IcebergObjectsByType sends the Iceberg
// table type filter and an ancestor filter, and that the response unmarshals
// into an IcebergObject, including the region and the parent taken from the
// physical path.
func TestIcebergObjectsByType(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	const response = `{
		"data": {
			"result": {
				"descendantConnection": {
					"count": 1,
					"nodes": [{
						"id": "11111111-1111-1111-1111-111111111111",
						"name": "orders",
						"objectType": "GLUE_ICEBERG_TABLE",
						"region": "US_EAST_1",
						"physicalPath": [{
							"fid": "22222222-2222-2222-2222-222222222222",
							"name": "example-account",
							"objectType": "AwsNativeAccount"
						}, {
							"fid": "33333333-3333-3333-3333-333333333333",
							"name": "us-east-1",
							"objectType": "GLUE_ICEBERG_CATALOG"
						}, {
							"fid": "44444444-4444-4444-4444-444444444444",
							"name": "sales",
							"objectType": "GLUE_ICEBERG_DATABASE"
						}]
					}],
					"pageInfo": {"endCursor": "", "hasNextPage": false}
				}
			}
		}
	}`

	var gotFilter []Filter
	var gotTypeFilter []ObjectType
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var decoded struct {
			Variables struct {
				Filter     []Filter     `json:"filter"`
				TypeFilter []ObjectType `json:"typeFilter"`
			} `json:"variables"`
		}
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		gotFilter = decoded.Variables.Filter
		gotTypeFilter = decoded.Variables.TypeFilter
		io.WriteString(w, response)
	}))
	defer srv.Close()

	catalogID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	tables, err := IcebergObjectsByType[GlueIcebergTable](ctx, Wrap(graphql.NewTestClient(srv)), AncestorFilter(catalogID))
	if err != nil {
		t.Fatalf("IcebergObjectsByType failed: %v", err)
	}

	if len(gotTypeFilter) != 1 || gotTypeFilter[0] != "GLUE_ICEBERG_TABLE" {
		t.Errorf("type filter: got %v, want [GLUE_ICEBERG_TABLE]", gotTypeFilter)
	}
	if len(gotFilter) != 1 || gotFilter[0].Field != "ANCESTOR_ID" ||
		len(gotFilter[0].Texts) != 1 || gotFilter[0].Texts[0] != catalogID.String() {
		t.Errorf("filter: got %+v, want ANCESTOR_ID=[%s]", gotFilter, catalogID)
	}

	if len(tables) != 1 {
		t.Fatalf("got %d tables, want 1", len(tables))
	}
	table := tables[0]
	if table.Name != "orders" {
		t.Errorf("name: got %q", table.Name)
	}
	if table.Region.Region != awsregions.RegionUsEast1 {
		t.Errorf("region: got %v, want us-east-1", table.Region)
	}
	parent, ok := table.Parent()
	if !ok {
		t.Fatal("parent: not found")
	}
	if parent.ID != uuid.MustParse("44444444-4444-4444-4444-444444444444") || parent.ObjectType != "GLUE_ICEBERG_DATABASE" {
		t.Errorf("parent: got %+v, want the sales database", parent)
	}
}

// TestIcebergObjectWithoutParent verifies that Parent reports no parent for an
// object with an empty physical path.
func TestIcebergObjectWithoutParent(t *testing.T) {
	if _, ok := (IcebergObject{}).Parent(); ok {
		t.Error("parent: got one for an empty physical path")
	}
}
