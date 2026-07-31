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

package devops

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GitHubOrganization represents a GitHub organization with the curated fields
// exposed by the SDK. GitHub organizations cannot be onboarded through the SDK;
// they are read-only.
type GitHubOrganization struct {
	sla.HierarchyObject
	NativeID               string                  `json:"nativeId"`
	ConnectionStatus       ConnectionStatus        `json:"connectionStatus"`
	RepoHostType           HostType                `json:"repoHostType"`
	DevOpsOrgType          OrgType                 `json:"devOpsOrgType"`
	RepoCount              int                     `json:"repoCount"`
	LastRefreshTime        *time.Time              `json:"lastRefreshTime"`
	OrgURL                 string                  `json:"orgUrl"`
	BackupLocation         *BackupLocation         `json:"backupLocation"`
	Exocompute             *CloudNativeExocompute  `json:"exocompute"`
	RubrikHostedExocompute *RubrikHostedExocompute `json:"rubrikHostedExocompute"`
}

// GitHubOrganizations returns all GitHub organizations under the specified
// ancestor. Pass hierarchy.GitHubRoot as the ancestor ID to enumerate every
// organization in the account; use queryType to select CHILDREN (one level) or
// DESCENDANTS (the whole subtree). Pass zero or more filters to narrow the
// results server-side, e.g. hierarchy.Filter{Field: "NAME_EXACT_MATCH", Texts:
// []string{name}}.
func GitHubOrganizations(ctx context.Context, gql *graphql.Client, queryType QueryType, ancestorID string, filters ...hierarchy.Filter) ([]GitHubOrganization, error) {
	gql.Log().Print(log.Trace)

	if filters == nil {
		filters = []hierarchy.Filter{}
	}
	var cursor string
	var orgs []GitHubOrganization
	for {
		query := githubOrganizationsQuery
		buf, err := gql.Request(ctx, query, struct {
			First      int                `json:"first"`
			After      string             `json:"after,omitempty"`
			QueryType  QueryType          `json:"queryType"`
			AncestorID string             `json:"ancestorId"`
			Filter     []hierarchy.Filter `json:"filter"`
		}{First: 100, After: cursor, QueryType: queryType, AncestorID: ancestorID, Filter: filters})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []GitHubOrganization `json:"nodes"`
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

		orgs = append(orgs, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return orgs, nil
}

// GitHubRepository represents a GitHub repository with the curated fields
// exposed by the SDK. The repository is the snappable object.
type GitHubRepository struct {
	sla.HierarchyObject
	OrgID   uuid.UUID `json:"orgId"`
	OrgName string    `json:"orgName"`
	Size    int64     `json:"size"`
}

// GitHubRepositories returns all GitHub repositories under the specified
// ancestor (typically an organization ID). Use queryType to select CHILDREN or
// DESCENDANTS. Pass zero or more filters to narrow the results server-side, e.g.
// hierarchy.Filter{Field: "NAME_EXACT_MATCH", Texts: []string{name}}.
func GitHubRepositories(ctx context.Context, gql *graphql.Client, queryType QueryType, ancestorID string, filters ...hierarchy.Filter) ([]GitHubRepository, error) {
	gql.Log().Print(log.Trace)

	if filters == nil {
		filters = []hierarchy.Filter{}
	}
	var cursor string
	var repos []GitHubRepository
	for {
		query := githubRepositoriesQuery
		buf, err := gql.Request(ctx, query, struct {
			First      int                `json:"first"`
			After      string             `json:"after,omitempty"`
			QueryType  QueryType          `json:"queryType"`
			AncestorID string             `json:"ancestorId"`
			Filter     []hierarchy.Filter `json:"filter"`
		}{First: 100, After: cursor, QueryType: queryType, AncestorID: ancestorID, Filter: filters})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []GitHubRepository `json:"nodes"`
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

		repos = append(repos, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return repos, nil
}
