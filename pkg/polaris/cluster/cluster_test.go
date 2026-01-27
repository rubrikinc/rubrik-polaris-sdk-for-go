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

package cluster

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

// newTestClient creates a polaris.Client for testing with the given HTTP test server.
func newTestClient(srv *httptest.Server) *polaris.Client {
	return &polaris.Client{
		GQL: graphql.NewTestClient(srv),
	}
}

func TestCanIgnoreClusterRemovalPrechecks(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/can_ignore_cluster_removal_prechecks_response.json.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if err := tmpl.Execute(w, struct {
			Disconnected   bool
			IgnorePrecheck bool
			AirGapped      bool
		}{Disconnected: true, IgnorePrecheck: true, AirGapped: false}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	clusterID := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	prechecks, err := api.CanIgnoreClusterRemovalPrechecks(ctx, clusterID)
	if err != nil {
		t.Fatal(err)
	}
	if !prechecks.Disconnected {
		t.Error("expected Disconnected to be true")
	}
	if !prechecks.IgnorePrecheck {
		t.Error("expected IgnorePrecheck to be true")
	}
}

func TestRemoveCDMCluster(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/remove_cdm_cluster_response.json.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if err := tmpl.Execute(w, struct{ Result bool }{Result: true}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	clusterID := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	result, err := api.RemoveCDMCluster(ctx, clusterID, false, 30)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Error("expected result to be true")
	}
}

func TestRemoveClusterSuccess(t *testing.T) {
	templates := template.Must(template.ParseFiles(
		"testdata/can_ignore_cluster_removal_prechecks_response.json.tmpl",
		"testdata/cluster_rcv_locations_response.json.tmpl",
		"testdata/verify_sla_with_replication_to_cluster_response.json.tmpl",
		"testdata/all_cluster_global_slas_response.json.tmpl",
		"testdata/remove_cdm_cluster_response.json.tmpl",
	))

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Responses returned in sequence based on query type
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}

		var payload struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		switch {
		case strings.Contains(payload.Query, "canIgnoreClusterRemovalPrechecks"):
			if err := templates.ExecuteTemplate(w, "can_ignore_cluster_removal_prechecks_response.json.tmpl", struct {
				Disconnected   bool
				IgnorePrecheck bool
				AirGapped      bool
			}{Disconnected: true, IgnorePrecheck: true, AirGapped: false}); err != nil {
				cancel(err)
			}
		case strings.Contains(payload.Query, "clusterRcvLocations"):
			if err := templates.ExecuteTemplate(w, "cluster_rcv_locations_response.json.tmpl", struct {
				Locations []struct{ ID, Name string }
			}{Locations: nil}); err != nil {
				cancel(err)
			}
		case strings.Contains(payload.Query, "verifySlaWithReplicationToCluster"):
			if err := templates.ExecuteTemplate(w, "verify_sla_with_replication_to_cluster_response.json.tmpl", struct {
				IsActiveSLA bool
			}{IsActiveSLA: false}); err != nil {
				cancel(err)
			}
		case strings.Contains(payload.Query, "allClusterGlobalSlas"):
			if err := templates.ExecuteTemplate(w, "all_cluster_global_slas_response.json.tmpl", struct {
				SLAs []struct{ ID, Name string }
			}{SLAs: nil}); err != nil {
				cancel(err)
			}
		case strings.Contains(payload.Query, "removeCdmCluster"):
			if err := templates.ExecuteTemplate(w, "remove_cdm_cluster_response.json.tmpl", struct {
				Result bool
			}{Result: true}); err != nil {
				cancel(err)
			}
		}
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	clusterID := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	info, err := api.RemoveCluster(ctx, clusterID, false, 30)
	if err != nil {
		t.Fatal(err)
	}

	if !info.Prechecks.Disconnected {
		t.Error("expected Disconnected to be true")
	}
	if info.BlockingConditions {
		t.Error("expected no blocking conditions")
	}
	if info.ForceRemovalEligible {
		t.Error("expected force removal not to be eligible (no blocking conditions)")
	}
}

func TestRemoveClusterForceRemovalNotEligibleAirGapped(t *testing.T) {
	templates := template.Must(template.ParseFiles(
		"testdata/can_ignore_cluster_removal_prechecks_response.json.tmpl",
		"testdata/cluster_rcv_locations_response.json.tmpl",
		"testdata/verify_sla_with_replication_to_cluster_response.json.tmpl",
		"testdata/all_cluster_global_slas_response.json.tmpl",
	))

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}

		var payload struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		switch {
		case strings.Contains(payload.Query, "canIgnoreClusterRemovalPrechecks"):
			if err := templates.ExecuteTemplate(w, "can_ignore_cluster_removal_prechecks_response.json.tmpl", struct {
				Disconnected   bool
				IgnorePrecheck bool
				AirGapped      bool
			}{Disconnected: true, IgnorePrecheck: true, AirGapped: true}); err != nil {
				cancel(err)
			}
		case strings.Contains(payload.Query, "clusterRcvLocations"):
			if err := templates.ExecuteTemplate(w, "cluster_rcv_locations_response.json.tmpl", struct {
				Locations []struct{ ID, Name string }
			}{Locations: []struct{ ID, Name string }{{ID: "a48e7ad0-7b86-4c96-b6ba-97eb6a82f765", Name: "RCV1"}}}); err != nil {
				cancel(err)
			}
		case strings.Contains(payload.Query, "verifySlaWithReplicationToCluster"):
			if err := templates.ExecuteTemplate(w, "verify_sla_with_replication_to_cluster_response.json.tmpl", struct {
				IsActiveSLA bool
			}{IsActiveSLA: true}); err != nil {
				cancel(err)
			}
		case strings.Contains(payload.Query, "allClusterGlobalSlas"):
			if err := templates.ExecuteTemplate(w, "all_cluster_global_slas_response.json.tmpl", struct {
				SLAs []struct{ ID, Name string }
			}{SLAs: []struct{ ID, Name string }{{ID: "d48e7ad0-7b86-4c96-b6ba-97eb6a82f765", Name: "SLA1"}}}); err != nil {
				cancel(err)
			}
		}
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	clusterID := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	_, err := api.RemoveCluster(ctx, clusterID, true, 30)
	if err == nil {
		t.Fatal("expected error for air-gapped cluster force removal")
	}
	if !strings.Contains(err.Error(), "air-gapped") {
		t.Errorf("unexpected error: %v", err)
	}
}
