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

package cloudapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlcloudapp "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudapp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlaws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// TestCreateAWSApplicationInvalidParams verifies that parameters RSC would
// accept or reject late are rejected without a request being sent to RSC.
func TestCreateAWSApplicationInvalidParams(t *testing.T) {
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		t.Error("unexpected request to RSC")
	}))
	defer srv.Close()

	accountID := uuid.MustParse("22bf439e-9922-486e-a32c-ae9009987924")
	valid := gqlcloudapp.CreateAWSApplicationParams{
		CloudAccountID:       accountID,
		TargetCloudAccountID: uuid.MustParse("ce09d0e4-d0a5-4201-9db2-97f0b6644514"),
		Name:                 "app",
		Region:               gqlaws.RegionUsEast1.ToRegionEnum(),
		MatchingTags:         []core.Tag{{Key: "Version", Value: "1.0"}},
	}

	unknownRegion := valid
	unknownRegion.Region = gqlaws.RegionUnknown.ToRegionEnum()
	sameAccount := valid
	sameAccount.TargetCloudAccountID = accountID
	noSelection := valid
	noSelection.MatchingTags = nil

	for name, params := range map[string]gqlcloudapp.CreateAWSApplicationParams{
		"UnknownRegion": unknownRegion,
		"SameAccount":   sameAccount,
		"NoSelection":   noSelection,
	} {
		if _, err := WrapGQL(graphql.NewTestClient(srv)).CreateAWSApplication(t.Context(), params); err == nil {
			t.Errorf("%s: expected CreateAWSApplication to fail", name)
		}
	}
}

// TestRemoveApplicationDeleteSnapshots verifies the order of the requests made
// when removing an application together with its snapshots: the application
// is set to Do Not Protect with its existing snapshots expiring immediately,
// its unmanaged snapshots are deleted, the config snapshots are polled until
// RSC reports none, and only then is the application deleted.
func TestRemoveApplicationDeleteSnapshots(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	appID := uuid.MustParse("c7ec46ba-d419-4a9d-88e2-26665e1479db")
	var operations []string
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Query     string                     `json:"query"`
			Variables map[string]json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			cancel(err)
			return
		}
		name := graphql.QueryName(body.Query)
		operations = append(operations, name)

		var response string
		switch {
		case name == "assignSla":
			if got := string(body.Variables["slaDomainAssignType"]); got != `"doNotProtect"` {
				t.Errorf("invalid assign type: %s", got)
			}
			if got := string(body.Variables["existingSnapshotRetention"]); got != `"EXPIRE_IMMEDIATELY"` {
				t.Errorf("invalid existing snapshot retention: %s", got)
			}
			if got := string(body.Variables["objectIds"]); got != `["`+appID.String()+`"]` {
				t.Errorf("invalid object IDs: %s", got)
			}
			response = `{"data":{"result":{"success":true}}}`
		case name == "deleteSnapshotsOfUnmanagedObjects":
			response = `{"data":{"result":{"success":true}}}`
		case name == "cloudNativeApplicationSnapshots":
			response = `{"data":null,"errors":[{"message":"NOT_FOUND: no config snapshot found for application",` +
				`"extensions":{"code":404}}]}`
		case name == "deleteCloudNativeApplication":
			response = `{"data":{"result":null}}`
		default:
			t.Errorf("unexpected operation: %s", name)
			response = `{"data":null}`
		}
		if _, err := w.Write([]byte(response)); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	if err := WrapGQL(graphql.NewTestClient(srv)).RemoveApplication(ctx, appID, true); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"assignSla",
		"deleteSnapshotsOfUnmanagedObjects",
		"cloudNativeApplicationSnapshots",
		"deleteCloudNativeApplication",
	}
	if !slices.Equal(operations, want) {
		t.Errorf("invalid operation order: %v", operations)
	}
}

// TestCreateAWSApplicationChecks verifies that an application isn't created
// when RSC reports the name as taken, since RSC's create accepts duplicate
// names, or when the cloud account's CLOUD_NATIVE_CONFIG_PROTECTION feature
// isn't connected or doesn't cover the application region, since RSC then
// persists the application without returning its ID.
func TestCreateAWSApplicationChecks(t *testing.T) {
	accountID := uuid.MustParse("22bf439e-9922-486e-a32c-ae9009987924")
	tests := []struct {
		name       string
		nameStatus string
		status     string
		regions    string
	}{
		{name: "NameTaken", nameStatus: "APP_NAME_NON_UNIQUE", status: "CONNECTED", regions: `["US_EAST_1"]`},
		{name: "NotConnected", nameStatus: "APP_NAME_VALID", status: "CONNECTING", regions: `["US_EAST_1"]`},
		{name: "OtherRegion", nameStatus: "APP_NAME_VALID", status: "CONNECTED", regions: `["EU_WEST_1"]`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				var body struct {
					Query string `json:"query"`
				}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					cancel(err)
					return
				}
				var response string
				switch name := graphql.QueryName(body.Query); name {
				case "isValidCloudNativeApplicationName":
					response = `{"data":{"result":{"status":"` + tc.nameStatus + `"}}}`
				case "allAwsCloudAccountsWithFeatures":
					response = `{"data":{"result":[{"awsCloudAccount":{"cloudType":"STANDARD",` +
						`"id":"` + accountID.String() + `","nativeId":"123456789012","accountName":"source","message":""},` +
						`"featureDetails":[{"feature":"CLOUD_NATIVE_CONFIG_PROTECTION","permissionsGroups":[],` +
						`"awsRegions":` + tc.regions + `,"roleArn":"","stackArn":"","status":"` + tc.status + `",` +
						`"mappedAccounts":[],"roleChainingDetails":null}],"roleChainingAccount":null}]}}`
				default:
					t.Errorf("unexpected operation: %s", name)
					return
				}
				if _, err := w.Write([]byte(response)); err != nil {
					cancel(err)
				}
			}))
			defer srv.Close()

			_, err := WrapGQL(graphql.NewTestClient(srv)).CreateAWSApplication(ctx, gqlcloudapp.CreateAWSApplicationParams{
				CloudAccountID:       accountID,
				TargetCloudAccountID: uuid.MustParse("ce09d0e4-d0a5-4201-9db2-97f0b6644514"),
				Name:                 "app",
				Region:               gqlaws.RegionUsEast1.ToRegionEnum(),
				MatchingTags:         []core.Tag{{Key: "Version", Value: "1.0"}},
			})
			if err == nil {
				t.Fatal("expected CreateAWSApplication to fail")
			}
		})
	}
}

// TestApplicationByID verifies that an application is only reported as found
// when it's in the application list, which doesn't return archived
// applications, and that the list is filtered by the application's own name
// and cloud account and matched by ID, since the name filter matches
// substrings.
func TestApplicationByID(t *testing.T) {
	appID := uuid.MustParse("c7ec46ba-d419-4a9d-88e2-26665e1479db")
	accountID := uuid.MustParse("22bf439e-9922-486e-a32c-ae9009987924")
	otherID := uuid.MustParse("9a8a7c63-0ae4-4be8-b0a2-1ab8b28a4f45")
	app := `{"data":{"result":{"id":"` + appID.String() + `","name":"web","cloudNativeAccountId":"` +
		accountID.String() + `","targetCloudNativeAccountId":"ce09d0e4-d0a5-4201-9db2-97f0b6644514",` +
		`"region":{"region":{"awsRegion":"US_EAST_1"}},"matchingTags":{"matchingTagsList":[]},` +
		`"resourceInclusionList":["arn:aws:lambda:us-east-1:123456789012:function:my-function"],` +
		`"resourceExclusionList":[],"refreshStatus":"APP_REFRESH_SUCCESS","protectionStatus":"PROTECTION_STATUS_UNPROTECTED"}}}`
	list := func(nodes string) string {
		return `{"data":{"result":{"nodes":[` + nodes + `],"pageInfo":{"endCursor":"","hasNextPage":false}}}}`
	}

	tests := []struct {
		name     string
		read     string
		list     string
		notFound bool
	}{
		{name: "Live", read: app, list: list(`{"id":"` + appID.String() + `","name":"web"}`)},
		{name: "Deleted", read: app, list: list(``), notFound: true},
		{name: "OtherNameMatch", read: app, list: list(`{"id":"` + otherID.String() + `","name":"web-store"}`), notFound: true},
		{name: "Unknown", read: `{"data":{"result":null},"errors":[{"message":"Objects are not authorized.",` +
			`"path":["result"],"extensions":{"code":403}}]}`, notFound: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				var body struct {
					Query     string                     `json:"query"`
					Variables map[string]json.RawMessage `json:"variables"`
				}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					cancel(err)
					return
				}
				var response string
				switch name := graphql.QueryName(body.Query); name {
				case "cloudNativeApplication":
					response = tc.read
				case "cloudNativeApplications":
					if got := string(body.Variables["applicationName"]); got != `"web"` {
						t.Errorf("invalid name filter: %s", got)
					}
					if got := string(body.Variables["cloudNativeAccountIds"]); got != `["`+accountID.String()+`"]` {
						t.Errorf("invalid cloud account filter: %s", got)
					}
					response = tc.list
				default:
					t.Errorf("unexpected operation: %s", name)
					return
				}
				if _, err := w.Write([]byte(response)); err != nil {
					cancel(err)
				}
			}))
			defer srv.Close()

			got, err := WrapGQL(graphql.NewTestClient(srv)).ApplicationByID(ctx, appID)
			if tc.notFound {
				if !errors.Is(err, graphql.ErrNotFound) {
					t.Fatalf("expected graphql.ErrNotFound, got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != appID || len(got.ResourceARNs) != 1 {
				t.Errorf("invalid application: %s %v", got.ID, got.ResourceARNs)
			}
		})
	}
}
