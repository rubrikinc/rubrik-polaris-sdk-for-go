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
	"fmt"
	"time"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ClosestSnapshotParams holds the parameters for a closest snapshot query.
// Either BeforeTime or AfterTime must be set, but not both. It is an error to
// supply both or neither. WorkloadID must be non-empty.
type ClosestSnapshotParams struct {
	WorkloadID         string
	BeforeTime         *time.Time
	AfterTime          *time.Time
	ExcludeQuarantined bool
	ExcludeAnomalous   bool
}

// ClosestSnapshotResult holds the result of a closest snapshot query for a
// single workload. Snapshot may be nil if no snapshot was found near the
// requested point in time.
type ClosestSnapshotResult struct {
	WorkloadID string `json:"snappableId"`
	Snapshot   *struct {
		ID   string    `json:"id"`
		Date time.Time `json:"date"`
	} `json:"snapshot"`
}

// ClosestSnapshot returns the closest snapshot to the given point in time for
// the specified workload.
func (a API) ClosestSnapshot(ctx context.Context, params ClosestSnapshotParams) (ClosestSnapshotResult, error) {
	a.log.Print(log.Trace)

	if params.WorkloadID == "" {
		return ClosestSnapshotResult{}, fmt.Errorf("WorkloadID must be non-empty")
	}
	if params.BeforeTime == nil && params.AfterTime == nil {
		return ClosestSnapshotResult{}, fmt.Errorf("either BeforeTime or AfterTime must be set")
	}
	if params.BeforeTime != nil && params.AfterTime != nil {
		return ClosestSnapshotResult{}, fmt.Errorf("BeforeTime and AfterTime are mutually exclusive")
	}

	query := closestSnapshotQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		WorkloadIDs        []string   `json:"snappableIds"`
		BeforeTime         *time.Time `json:"beforeTime,omitempty"`
		AfterTime          *time.Time `json:"afterTime,omitempty"`
		ExcludeQuarantined bool       `json:"excludeQuarantined"`
		ExcludeAnomalous   bool       `json:"excludeAnomalous"`
	}{
		WorkloadIDs:        []string{params.WorkloadID},
		BeforeTime:         params.BeforeTime,
		AfterTime:          params.AfterTime,
		ExcludeQuarantined: params.ExcludeQuarantined,
		ExcludeAnomalous:   params.ExcludeAnomalous,
	})
	if err != nil {
		return ClosestSnapshotResult{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []struct {
				ClosestSnapshotResult
				Error string `json:"error"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ClosestSnapshotResult{}, graphql.UnmarshalError(query, err)
	}

	if len(payload.Data.Result) == 0 {
		return ClosestSnapshotResult{}, fmt.Errorf("no result returned for workload %s", params.WorkloadID)
	}

	result := payload.Data.Result[0]
	if result.Error != "" {
		return ClosestSnapshotResult{}, fmt.Errorf("closest snapshot query failed for workload %s: %s", params.WorkloadID, result.Error)
	}

	return result.ClosestSnapshotResult, nil
}
