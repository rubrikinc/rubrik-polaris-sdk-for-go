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

package cloudcluster

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// TestAllAzureCdmVersionsLocation verifies that the region is sent to
// azureCcCdmVersions as the native Azure location name, which RSC passes
// through to Azure. This includes regions without an AzureCloudAccountRegion
// enum value.
func TestAllAzureCdmVersionsLocation(t *testing.T) {
	cloudAccountID := uuid.New()

	for _, test := range []struct {
		region   azure.Region
		location string
	}{
		{azure.RegionGermanyWestCentral, "germanywestcentral"},
		{azure.RegionEastUS2, "eastus2"},
		{azure.RegionNewZealandNorth, "newzealandnorth"},
	} {
		t.Run(test.location, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			var gotLocation string
			srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				var body struct {
					Variables struct {
						Location string `json:"location"`
					} `json:"variables"`
				}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					cancel(err)
					return
				}
				gotLocation = body.Variables.Location
				_, _ = w.Write([]byte(`{"data":{"result":[]}}`))
			}))
			defer srv.Close()

			if _, err := Wrap(graphql.NewTestClient(srv)).AllAzureCdmVersions(ctx, cloudAccountID, test.region); err != nil {
				t.Fatal(err)
			}
			if gotLocation != test.location {
				t.Errorf("got location %q want %q", gotLocation, test.location)
			}
		})
	}
}

// TestAllAzureCdmVersionsAllRegions verifies that every region known to the
// SDK, parsed from its name the same way the terraform provider does, is sent
// to azureCcCdmVersions as its native Azure location name.
func TestAllAzureCdmVersionsAllRegions(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	var gotLocation string
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables struct {
				Location string `json:"location"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			cancel(err)
			return
		}
		gotLocation = body.Variables.Location
		_, _ = w.Write([]byte(`{"data":{"result":[]}}`))
	}))
	defer srv.Close()

	api := Wrap(graphql.NewTestClient(srv))
	cloudAccountID := uuid.New()
	for _, name := range azure.AllRegionNames() {
		gotLocation = ""
		if _, err := api.AllAzureCdmVersions(ctx, cloudAccountID, azure.RegionFromName(name)); err != nil {
			t.Errorf("region %s: %v", name, err)
			continue
		}
		if gotLocation != name {
			t.Errorf("region %s: got location %q want %q", name, gotLocation, name)
		}
	}
}

// TestAllAzureCdmVersionsUnknownRegion verifies that an unknown region is
// rejected without sending a request to RSC.
func TestAllAzureCdmVersionsUnknownRegion(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		cancel(errors.New("unexpected request to RSC"))
	}))
	defer srv.Close()

	cloudAccountID := uuid.New()
	if _, err := Wrap(graphql.NewTestClient(srv)).AllAzureCdmVersions(ctx, cloudAccountID, azure.RegionUnknown); err == nil {
		t.Fatal("expected error for unknown region")
	}
}
