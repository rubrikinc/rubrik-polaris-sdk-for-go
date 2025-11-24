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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeProject represents an RSC native project. NativeProjects are connected
// to CloudAccount through the CloudAccountID and NativeID fields.
type NativeProject struct {
	ID               uuid.UUID      `json:"id"`
	CloudAccountID   uuid.UUID      `json:"cloudAccountId"`
	Name             string         `json:"name"`
	NativeID         string         `json:"nativeId"`
	NativeName       string         `json:"nativeName"`
	ProjectNumber    string         `json:"projectNumber"`
	OrganizationName string         `json:"organizationName"`
	Assignment       sla.Assignment `json:"slaAssignment"`
	Configured       sla.DomainRef  `json:"configuredSlaDomain"`
	Effective        sla.DomainRef  `json:"effectiveSlaDomain"`
}

// NativeProject returns the native project with the specified RSC native
// project ID.
func (a API) NativeProject(ctx context.Context, nativeProjectID uuid.UUID) (NativeProject, error) {
	a.log.Print(log.Trace)

	query := gcpNativeProjectQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID uuid.UUID `json:"fid"`
	}{ID: nativeProjectID})
	if err != nil {
		return NativeProject{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Account NativeProject `json:"gcpNativeProject"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return NativeProject{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Account, nil
}

// NativeProjects returns the native projects matching the specified filter.
// The filter can be used to search for a substring in project name or number.
func (a API) NativeProjects(ctx context.Context, filter string) ([]NativeProject, error) {
	a.log.Print(log.Trace)

	var accounts []NativeProject
	var cursor string
	for {
		query := gcpNativeProjectsQuery
		buf, err := a.GQL.Request(ctx, query, struct {
			After  string `json:"after,omitempty"`
			Filter string `json:"filter"`
		}{After: cursor, Filter: filter})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

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
			return nil, graphql.UnmarshalError(query, err)
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
// with the specified RSC native project ID. If deleteSnapshots is true the
// snapshots are deleted. Returns the RSC task chain ID.
func (a API) NativeDisableProject(ctx context.Context, nativeProjectID uuid.UUID, deleteSnapshots bool) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := gcpNativeDisableProjectQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID              uuid.UUID `json:"projectId"`
		DeleteSnapshots bool      `json:"shouldDeleteNativeSnapshots"`
	}{ID: nativeProjectID, DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Query struct {
				JobID uuid.UUID `json:"jobId"`
				Error string    `json:"error"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	if payload.Data.Query.Error != "" {
		return uuid.Nil, graphql.ResponseError(query, errors.New(payload.Data.Query.Error))
	}
	return payload.Data.Query.JobID, nil
}
