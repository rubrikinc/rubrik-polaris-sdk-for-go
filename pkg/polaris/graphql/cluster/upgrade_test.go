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
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

func TestCDMInfoFilterOmitsEmptyFields(t *testing.T) {
	enabled := true
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	in := CDMInfoFilter{
		ID:               []uuid.UUID{id},
		UpgradeScheduled: &enabled,
	}
	buf, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := []string{"id", "upgradeScheduled"}
	keys := make([]string, 0, len(got))
	for k := range got {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	if !slices.Equal(keys, want) {
		t.Fatalf("unexpected keys: got %v, want %v", keys, want)
	}
	if ids, _ := got["id"].([]any); len(ids) != 1 || ids[0] != id.String() {
		t.Fatalf("unexpected id: %v", got["id"])
	}
	if got["upgradeScheduled"] != true {
		t.Fatalf("unexpected upgradeScheduled: %v", got["upgradeScheduled"])
	}
}

func TestCDMInfoUnmarshal(t *testing.T) {
	src := `{
        "clusterUuid": "11111111-2222-3333-4444-555555555555",
        "version": "9.2.1",
        "versionStatus": "UpToDate",
        "clusterJobStatus": "Idle",
        "clusterStatus": {"message": "Healthy", "status": "READY"},
        "currentStateProgress": 0.5,
        "downloadedVersion": "9.2.2",
        "fastUpgradePreferred": false,
        "isRuSupported": true,
        "overallProgress": 0.75,
        "scheduleUpgradeAt": "2026-06-01T01:02:03Z",
        "upgradeRecommendationInfo": {
            "recommendation": "9.2.2",
            "nextReleaseRecommendation": "9.3.0",
            "upgradability": ["9.2.2","9.3.0"]
        },
        "lastUpgradeDuration": {"clusterUuid": "11111111-2222-3333-4444-555555555555", "fastUpgradeDuration": 600, "rollingUpgradeDuration": 1800}
    }`
	var got CDMInfo
	if err := json.Unmarshal([]byte(src), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Version != "9.2.1" || got.VersionStatus != "UpToDate" || got.DownloadedVersion != "9.2.2" {
		t.Fatalf("unexpected basic fields: %+v", got)
	}
	if !got.IsRUSupported {
		t.Fatalf("expected IsRUSupported true")
	}
	if got.ScheduleUpgradeAt == nil || !got.ScheduleUpgradeAt.Equal(time.Date(2026, 6, 1, 1, 2, 3, 0, time.UTC)) {
		t.Fatalf("unexpected scheduleUpgradeAt: %+v", got.ScheduleUpgradeAt)
	}
	if got.UpgradeRecommendationInfo == nil || got.UpgradeRecommendationInfo.Recommendation != "9.2.2" {
		t.Fatalf("unexpected recommendation: %+v", got.UpgradeRecommendationInfo)
	}
	if got.LastUpgradeDuration == nil || got.LastUpgradeDuration.FastUpgradeDuration != 600 {
		t.Fatalf("unexpected lastUpgradeDuration: %+v", got.LastUpgradeDuration)
	}
}

// TestClusterWithUpgradesInfoExtrapolatesFilter asserts that the filter
// struct fields are sent as individual top-level variables in the request
// body (i.e. that the call extrapolates the input rather than nesting it).
func TestClusterWithUpgradesInfoExtrapolatesFilter(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	clusterID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if _, ok := body.Variables["upgradeFilter"]; ok {
			http.Error(w, "filter must be extrapolated, not passed as composite", http.StatusBadRequest)
			return
		}
		for _, key := range []string{"id", "first"} {
			if _, ok := body.Variables[key]; !ok {
				http.Error(w, fmt.Sprintf("missing variable %q", key), http.StatusBadRequest)
				return
			}
		}
		fmt.Fprintf(w, `{
            "data": {
                "result": {
                    "edges": [{
                        "cursor": "c1",
                        "node": {
                            "id": "%s",
                            "name": "cluster-a",
                            "cdmUpgradeInfo": {
                                "clusterUuid": "%s",
                                "version": "9.2.1"
                            }
                        }
                    }],
                    "pageInfo": {"startCursor": "c1", "endCursor": "c1", "hasPreviousPage": false, "hasNextPage": false},
                    "count": 1
                }
            }
        }`, clusterID, clusterID)
	}))
	defer srv.Close()

	pageSize := 10
	got, err := ClusterWithUpgradesInfo(ctx, graphql.NewTestClient(srv),
		&CDMInfoFilter{ID: []uuid.UUID{clusterID}},
		UpgradeInfoSortByClusterName,
		core.Pagination{First: &pageSize},
	)
	if err != nil {
		t.Fatalf("ClusterWithUpgradesInfo: %v", err)
	}
	if len(got.Details) != 1 || got.Details[0].ID != clusterID {
		t.Fatalf("unexpected details: %+v", got.Details)
	}
	if got.Details[0].CDMInfo == nil || got.Details[0].CDMInfo.Version != "9.2.1" {
		t.Fatalf("unexpected CDMInfo: %+v", got.Details[0].CDMInfo)
	}
	if got.Count != 1 {
		t.Fatalf("unexpected count: %d", got.Count)
	}
	if got.PageInfo.HasNextPage {
		t.Fatalf("expected HasNextPage false")
	}
}

// TestClusterWithUpgradesInfoNilFilter ensures a nil filter is accepted.
func TestClusterWithUpgradesInfoNilFilter(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{"data":{"result":{"edges":[],"pageInfo":{"hasNextPage":false},"count":0}}}`)
	}))
	defer srv.Close()

	got, err := ClusterWithUpgradesInfo(ctx, graphql.NewTestClient(srv), nil, "", core.Pagination{})
	if err != nil {
		t.Fatalf("ClusterWithUpgradesInfo: %v", err)
	}
	if len(got.Details) != 0 {
		t.Fatalf("expected 0 details, got %d", len(got.Details))
	}
}
