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
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/hierarchy"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// QueryType selects the hierarchy traversal mode for the list queries. It maps
// to the schema QueryType enum.
type QueryType string

const (
	// QueryTypeChildren returns only the direct children of the ancestor.
	QueryTypeChildren QueryType = "CHILDREN"
	// QueryTypeDescendants returns all descendants of the ancestor.
	QueryTypeDescendants QueryType = "DESCENDANTS"
)

// BackupLocation represents the backup location associated with a DevOps
// organization. Nil on the organization when no backup location is configured.
type BackupLocation struct {
	ID              uuid.UUID   `json:"id"`
	ArchivalGroupID uuid.UUID   `json:"archivalGroupId"`
	Name            string      `json:"name"`
	StorageType     StorageType `json:"storageType"`
	Region          struct {
		AzureRegion azure.RegionEnum `json:"azureRegion"`
	} `json:"cloudSpecificRegion"`
}

// CloudNativeExocompute represents the customer cloud-native exocompute
// associated with a DevOps organization. Nil on the organization when no
// cloud-native exocompute is configured.
type CloudNativeExocompute struct {
	ID       uuid.UUID `json:"id"`
	HostName string    `json:"hostName"`
	Region   struct {
		Region struct {
			AzureRegion azure.CommonRegionEnum `json:"azureRegion"`
		} `json:"region"`
	} `json:"region"`
}

// RubrikHostedExocompute represents the Rubrik-hosted exocompute associated
// with a DevOps organization. Nil on the organization when no Rubrik-hosted
// exocompute is configured.
type RubrikHostedExocompute struct {
	ExocomputeClusterID uuid.UUID    `json:"exocomputeClusterId"`
	Region              azure.Region `json:"region"`
}

// AzureOrganization represents an onboarded Azure DevOps organization with the
// curated fields exposed by the SDK.
type AzureOrganization struct {
	ID                      uuid.UUID               `json:"id"`
	NativeID                string                  `json:"nativeId"`
	TenantDomain            string                  `json:"tenantId"`
	TenantID                *uuid.UUID              `json:"tenantUuid"`
	Cloud                   gqlazure.Cloud          `json:"cloudType"`
	ConnectionStatus        ConnectionStatus        `json:"connectionStatus"`
	AuthenticationMechanism AuthMechanism           `json:"authenticationMechanism"`
	ClientID                string                  `json:"clientId"`
	RepoHostType            HostType                `json:"repoHostType"`
	DevOpsOrgType           OrgType                 `json:"devOpsOrgType"`
	ProjectCount            int                     `json:"projectCount"`
	RepoCount               int                     `json:"repoCount"`
	LastRefreshTime         *time.Time              `json:"lastRefreshTime"`
	Name                    string                  `json:"name"`
	ObjectType              string                  `json:"objectType"`
	BackupLocation          *BackupLocation         `json:"backupLocation"`
	CloudNativeExocompute   *CloudNativeExocompute  `json:"cloudNativeExocompute"`
	RubrikHostedExocompute  *RubrikHostedExocompute `json:"rubrikHostedExocompute"`
}

// applyReadWorkaround fills in the fields that RSC does not yet return on
// read. Each field is only set when absent, so once the API starts returning a
// value the response takes precedence. Remove this once the API returns all the
// fields.
func (o *AzureOrganization) applyReadWorkaround() {
	if o.Cloud == "" {
		o.Cloud = gqlazure.PublicCloud
	}
}

// AzureProject represents an Azure DevOps project with the curated fields
// exposed by the SDK.
type AzureProject struct {
	ID         uuid.UUID `json:"id"`
	NativeID   string    `json:"nativeId"`
	Name       string    `json:"name"`
	OrgID      uuid.UUID `json:"orgId"`
	OrgName    string    `json:"orgName"`
	URL        string    `json:"url"`
	RepoCount  int       `json:"repoCount"`
	ObjectType string    `json:"objectType"`
}

// AzureRepository represents an Azure DevOps repository with the curated fields
// exposed by the SDK. The repository is the snappable object.
type AzureRepository struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	OrgID       uuid.UUID `json:"orgId"`
	OrgName     string    `json:"orgName"`
	ProjectID   uuid.UUID `json:"projectId"`
	ProjectName string    `json:"projectName"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	ObjectType  string    `json:"objectType"`
}

// AzureOrganizations returns all Azure DevOps organizations under the specified
// ancestor. Pass hierarchy.AzureDevOpsRoot as the ancestor ID to enumerate every
// organization in the account; use queryType to select CHILDREN (one level) or
// DESCENDANTS (the whole subtree). Pass zero or more filters to narrow the
// results server-side, e.g. hierarchy.Filter{Field: "NAME_EXACT_MATCH", Texts:
// []string{name}}.
func AzureOrganizations(ctx context.Context, gql *graphql.Client, queryType QueryType, ancestorID string, filters ...hierarchy.Filter) ([]AzureOrganization, error) {
	gql.Log().Print(log.Trace)

	if filters == nil {
		filters = []hierarchy.Filter{}
	}
	var cursor string
	var orgs []AzureOrganization
	for {
		query := azureDevopsOrganizationsQuery
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
					Nodes    []AzureOrganization `json:"nodes"`
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

	// Workaround: fill in the fields RSC does not yet return on read.
	for i := range orgs {
		orgs[i].applyReadWorkaround()
	}

	return orgs, nil
}

// AzureProjects returns all Azure DevOps projects under the specified ancestor
// (typically an organization ID). Use queryType to select CHILDREN or
// DESCENDANTS. Pass zero or more filters to narrow the results server-side, e.g.
// hierarchy.Filter{Field: "NAME_EXACT_MATCH", Texts: []string{name}}.
func AzureProjects(ctx context.Context, gql *graphql.Client, queryType QueryType, ancestorID string, filters ...hierarchy.Filter) ([]AzureProject, error) {
	gql.Log().Print(log.Trace)

	if filters == nil {
		filters = []hierarchy.Filter{}
	}
	var cursor string
	var projects []AzureProject
	for {
		query := azureDevopsProjectsQuery
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
					Nodes    []AzureProject `json:"nodes"`
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

		projects = append(projects, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return projects, nil
}

// AzureRepositories returns all Azure DevOps repositories under the specified
// ancestor (typically an organization or project ID). Use queryType to select
// CHILDREN or DESCENDANTS. Pass zero or more filters to narrow the results
// server-side, e.g. hierarchy.Filter{Field: "NAME_EXACT_MATCH", Texts:
// []string{name}}.
func AzureRepositories(ctx context.Context, gql *graphql.Client, queryType QueryType, ancestorID string, filters ...hierarchy.Filter) ([]AzureRepository, error) {
	gql.Log().Print(log.Trace)

	if filters == nil {
		filters = []hierarchy.Filter{}
	}
	var cursor string
	var repos []AzureRepository
	for {
		query := azureDevopsRepositoriesQuery
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
					Nodes    []AzureRepository `json:"nodes"`
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
