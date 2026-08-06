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

package azure

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeSubscription represents an RSC native subscription. Native
// subscriptions are connected to cloud accounts through the CloudAccountID and
// NativeID fields.
type NativeSubscription struct {
	ID             uuid.UUID      `json:"id"`                        // RSC object ID
	CloudAccountID uuid.UUID      `json:"accountConnectionId"`       // RSC cloud account ID
	NativeID       uuid.UUID      `json:"azureSubscriptionNativeId"` // Azure subscription ID
	Name           string         `json:"name"`
	Status         string         `json:"azureSubscriptionStatus"`
	SLAAssignment  sla.Assignment `json:"slaAssignment"`
	Configured     sla.DomainRef  `json:"configuredSlaDomain"`
	Effective      sla.DomainRef  `json:"effectiveSlaDomain"`
}

// Deprecated: use NativeSubscriptionsByFilter instead.
func (a API) NativeSubscriptions(ctx context.Context, filter string) ([]NativeSubscription, error) {
	a.log.Print(log.Trace)

	var subscriptions []NativeSubscription
	var cursor string
	for {
		query := azureNativeSubscriptionsQuery
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
						Node NativeSubscription `json:"node"`
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
		for _, subscription := range payload.Data.Result.Edges {
			subscriptions = append(subscriptions, subscription.Node)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return subscriptions, nil
}

// NativeSubscriptionFilters holds the criteria used to filter native
// subscriptions. All non-empty fields are combined with logical AND. A zero
// value matches every subscription.
type NativeSubscriptionFilters struct {
	// EffectiveSLAIDs restricts results to subscriptions whose effective SLA
	// domain is one of the given IDs.
	EffectiveSLAIDs []string

	// NameSubstring restricts results to subscriptions whose name contains
	// this substring.
	NameSubstring string
}

// NativeSubscriptionsByFilter returns the native subscriptions matching the
// specified filters. Results are sorted by name in ascending order.
func (a API) NativeSubscriptionsByFilter(ctx context.Context, filters NativeSubscriptionFilters) ([]NativeSubscription, error) {
	a.log.Print(log.Trace)

	var subscriptions []NativeSubscription
	var cursor string
	for {
		// Note, the query defaults to only return native subscriptions for the
		// VM protection feature, therefore the query must always contain all
		// the support protection features.
		query := azureNativeSubscriptionsWithFilterQuery
		buf, err := a.GQL.Request(ctx, query, struct {
			After   string `json:"after,omitempty"`
			Filters any    `json:"filters"`
		}{After: cursor, Filters: struct {
			EffectiveSSLAFilter struct {
				EffectiveSLAIDs []string `json:"effectiveSlaIds,omitempty"`
			} `json:"effectiveSlaFilter,omitzero"`
			NameSubstringFilter struct {
				NameSubstring string `json:"nameSubstring,omitempty"`
			} `json:"nameSubstringFilter,omitzero"`
		}{
			EffectiveSSLAFilter: struct {
				EffectiveSLAIDs []string `json:"effectiveSlaIds,omitempty"`
			}{EffectiveSLAIDs: filters.EffectiveSLAIDs},
			NameSubstringFilter: struct {
				NameSubstring string `json:"nameSubstring,omitempty"`
			}{NameSubstring: filters.NameSubstring},
		}})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []NativeSubscription `json:"nodes"`
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
		subscriptions = append(subscriptions, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return subscriptions, nil
}

// NativeResourceGroup represents an Azure resource group surfaced by RSC.
type NativeResourceGroup struct {
	ID           string `json:"id"` // RSC object ID
	Name         string `json:"name"`
	Subscription struct {
		ID             string    `json:"id"`                  // RSC object ID
		CloudAccountID uuid.UUID `json:"accountConnectionId"` // RSC cloud account ID
		NativeID       uuid.UUID `json:"nativeId"`            // Azure subscription ID
		Name           string    `json:"name"`
	} `json:"azureSubscriptionDetails"`
	SLAAssignment sla.Assignment `json:"slaAssignment"`
	LogicalPath   []PathNode     `json:"logicalPath"`
	PhysicalPath  []PathNode     `json:"physicalPath"`
}

// PathNode is one step in the logical/physical hierarchy path for an RSC
// inventory object.
type PathNode struct {
	FID        string `json:"fid"`
	Name       string `json:"name"`
	ObjectType string `json:"objectType"`
}

// Deprecated: use NativeResourceGroupsByFilter instead.
func (a API) NativeResourceGroups(ctx context.Context, subscriptionIDs []uuid.UUID, nameSubstring string) ([]NativeResourceGroup, error) {
	a.log.Print(log.Trace)

	ids := make([]string, 0, len(subscriptionIDs))
	for _, id := range subscriptionIDs {
		ids = append(ids, id.String())
	}

	var groups []NativeResourceGroup
	var cursor string
	for {
		query := azureNativeResourceGroupsQuery
		buf, err := a.GQL.Request(ctx, query, struct {
			After           string   `json:"after,omitempty"`
			SubscriptionIDs []string `json:"subscriptionIds"`
			NameSubstring   string   `json:"nameSubstring"`
		}{After: cursor, SubscriptionIDs: ids, NameSubstring: nameSubstring})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Edges []struct {
						Node NativeResourceGroup `json:"node"`
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
		for _, edge := range payload.Data.Result.Edges {
			groups = append(groups, edge.Node)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return groups, nil
}

// ResourceGroupFilters holds the criteria used to filter native resource
// groups. All non-empty fields are combined with logical AND. A zero value
// matches every resource group.
type ResourceGroupFilters struct {
	// EffectiveSLAIDs restricts results to resource groups whose effective SLA
	// domain is one of the given IDs.
	EffectiveSLAIDs []string

	// NameSubstring restricts results to resource groups whose name contains
	// this substring.
	NameSubstring string

	// Regions restricts results to resource groups in one of the given
	// regions.
	Regions []azure.Region

	// SubscriptionIDs restricts results to resource groups belonging to one of
	// the given subscriptions. Note, the ID is the hierarchy/inventory ID.
	SubscriptionIDs []uuid.UUID
}

// NativeResourceGroupsByFilter returns the native resource groups matching the
// specified filters. Results are sorted by subscription name in ascending
// order.
//
// Note, if this function is called during an RSC subscription refresh, resource
// groups might be included more than once.
func (a API) NativeResourceGroupsByFilter(ctx context.Context, filters ResourceGroupFilters) ([]NativeResourceGroup, error) {
	a.log.Print(log.Trace)

	// Note nativeRegions must be nil when filter.Regions field is empty,
	// otherwise the GraphQL call fails.
	var nativeRegions []azure.NativeRegionEnum
	for _, region := range filters.Regions {
		nativeRegions = append(nativeRegions, region.ToNativeRegionEnum())
	}

	var groups []NativeResourceGroup
	var cursor string
	for {
		query := azureNativeResourceGroupsWithFilterQuery
		buf, err := a.GQL.Request(ctx, query, struct {
			After   string `json:"after,omitempty"`
			Filters any    `json:"filters,omitzero"`
		}{After: cursor, Filters: struct {
			EffectiveSlaFilter struct {
				EffectiveSlaIds []string `json:"effectiveSlaIds,omitempty"`
			} `json:"effectiveSlaFilter,omitzero"`
			NameSubstringFilter struct {
				NameSubstring string `json:"nameSubstring,omitempty"`
			} `json:"nameSubstringFilter,omitzero"`
			RegionFilter struct {
				Regions []azure.NativeRegionEnum `json:"regions,omitempty"`
			} `json:"regionFilter,omitzero"`
			SubscriptionFilter struct {
				SubscriptionIds []uuid.UUID `json:"subscriptionIds,omitempty"`
			} `json:"subscriptionFilter,omitzero"`
		}{
			EffectiveSlaFilter: struct {
				EffectiveSlaIds []string `json:"effectiveSlaIds,omitempty"`
			}{EffectiveSlaIds: filters.EffectiveSLAIDs},
			NameSubstringFilter: struct {
				NameSubstring string `json:"nameSubstring,omitempty"`
			}{NameSubstring: filters.NameSubstring},
			RegionFilter: struct {
				Regions []azure.NativeRegionEnum `json:"regions,omitempty"`
			}{Regions: nativeRegions},
			SubscriptionFilter: struct {
				SubscriptionIds []uuid.UUID `json:"subscriptionIds,omitempty"`
			}{SubscriptionIds: filters.SubscriptionIDs},
		}})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []NativeResourceGroup `json:"nodes"`
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
		groups = append(groups, payload.Data.Result.Nodes...)
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return groups, nil
}

// Deprecated: no replacement.
type ProtectionFeature string

const (
	// Deprecated: no replacement.
	SQLDB ProtectionFeature = "SQL_DB"

	// Deprecated: no replacement.
	SQLMI ProtectionFeature = "SQL_MI"

	// Deprecated: no replacement.
	VM ProtectionFeature = "VM"
)

// Deprecated: use StartDisableCloudAccountJob instead.
func (a API) StartDisableNativeSubscriptionProtectionJob(ctx context.Context, id uuid.UUID, feature ProtectionFeature, deleteSnapshots bool) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := startDisableAzureNativeSubscriptionProtectionJobQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID              uuid.UUID         `json:"azureSubscriptionRubrikId"`
		DeleteSnapshots bool              `json:"shouldDeleteNativeSnapshots"`
		Feature         ProtectionFeature `json:"azureNativeProtectionFeature"`
	}{ID: id, DeleteSnapshots: deleteSnapshots, Feature: feature})
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				JobID uuid.UUID `json:"jobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.JobID, nil
}
