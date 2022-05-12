// Copyright 2021 Rubrik, Inc.
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

package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeProject represents a Polaris native project. NativeProjects are
// connected to CloudAccounts through the NativeID field.
type NativeProject struct {
	ID               uuid.UUID          `json:"id"`
	Name             string             `json:"name"`
	NativeID         string             `json:"nativeId"`
	NativeName       string             `json:"nativeName"`
	ProjectNumber    string             `json:"projectNumber"`
	OrganizationName string             `json:"organizationName"`
	Assignment       core.SLAAssignment `json:"slaAssignment"`
	Configured       core.SLADomain     `json:"configuredSlaDomain"`
	Effective        core.SLADomain     `json:"effectiveSlaDomain"`
}

// NativeProject returns the native project with the specified Polaris native
// project id.
func (a API) NativeProject(ctx context.Context, id uuid.UUID) (NativeProject, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, gcpNativeProjectQuery, struct {
		ID uuid.UUID `json:"fid"`
	}{ID: id})
	if err != nil {
		return NativeProject{}, fmt.Errorf("failed to request NativeProject: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "gcpNativeProject(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Account NativeProject `json:"gcpNativeProject"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return NativeProject{}, fmt.Errorf("failed to unmarshal NativeProject: %v", err)
	}

	return payload.Data.Account, nil
}

// NativeProjects returns the native projects matching the specified filter.
// The filter can be used to search for a substring in project name or number.
func (a API) NativeProjects(ctx context.Context, filter string) ([]NativeProject, error) {
	a.GQL.Log().Print(log.Trace)

	var accounts []NativeProject
	var cursor string
	for {
		query := gcpNativeProjectsQuery
		if graphql.VersionOlderThan(a.Version, "master-46700", "v20220412") {
			query = gcpNativeProjectConnectionQuery
		}
		buf, err := a.GQL.Request(ctx, query, struct {
			After  string `json:"after,omitempty"`
			Filter string `json:"filter"`
		}{After: cursor, Filter: filter})
		if err != nil {
			return nil, fmt.Errorf("failed to request NativeProjects: %v", err)
		}

		a.GQL.Log().Printf(log.Debug, "%s(%q): %s", graphql.QueryName(query), filter, string(buf))

		var payload struct {
			Data struct {
				Result struct {
					Count int `json:"count"`
					Edges []struct {
						Node NativeProject `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal NativeProjects: %v", err)
		}
		for _, account := range payload.Data.Result.Edges {
			accounts = append(accounts, account.Node)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return accounts, nil
}

// NativeDisableProject starts a task chain job to disable the native project
// with the specified Polaris native project id. If deleteSnapshots is true the
// snapshots are deleted. Returns the Polaris task chain id.
func (a API) NativeDisableProject(ctx context.Context, id uuid.UUID, deleteSnapshots bool) (uuid.UUID, error) {
	a.GQL.Log().Print(log.Trace)

	query := gcpNativeDisableProjectQuery
	if graphql.VersionOlderThan(a.Version, "master-46700", "v20220412") {
		query = gcpNativeDisableProjectV0Query
	} else if graphql.VersionOlderThan(a.Version, "master-47076", "v20220426") {
		query = gcpNativeDisableProjectV1Query
	}
	buf, err := a.GQL.Request(ctx, query, struct {
		ID              uuid.UUID `json:"projectId"`
		DeleteSnapshots bool      `json:"shouldDeleteNativeSnapshots"`
	}{ID: id, DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request NativeDisableProject: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "%s(%q, %t): %s", graphql.QueryName(query), id, deleteSnapshots, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				JobID       uuid.UUID `json:"jobId"`
				TaskChainID uuid.UUID `json:"taskchainUuid"`
				Error       string    `json:"error"`
			} `json:"gcpNativeDisableProject"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal NativeDisableProject: %v", err)
	}

	if graphql.VersionOlderThan(a.Version, "master-46700", "v20220412") {
		return payload.Data.Query.TaskChainID, nil
	}
	if graphql.VersionOlderThan(a.Version, "master-47076", "v20220426") {
		return payload.Data.Query.JobID, nil
	}
	if payload.Data.Query.Error != "" {
		return uuid.Nil, errors.New(payload.Data.Query.Error)
	}
	return payload.Data.Query.JobID, nil
}
