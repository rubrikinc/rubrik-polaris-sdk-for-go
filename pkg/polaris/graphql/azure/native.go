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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/sla"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// NativeSubscription represents an RSC native subscription. NativeSubscriptions
// are connected to CloudAccount through the CloudAccountID and NativeID fields.
type NativeSubscription struct {
	ID             uuid.UUID      `json:"id"`
	CloudAccountID uuid.UUID      `json:"accountConnectionId"`
	Name           string         `json:"name"`
	NativeID       uuid.UUID      `json:"azureSubscriptionNativeId"`
	Status         string         `json:"azureSubscriptionStatus"`
	SLAAssignment  sla.Assignment `json:"slaAssignment"`
	Configured     sla.DomainRef  `json:"configuredSlaDomain"`
	Effective      sla.DomainRef  `json:"effectiveSlaDomain"`
}

// NativeSubscriptions returns the native subscriptions matching the specified
// filter. The filter can be used to search for a substring in the subscription
// name.
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

// PathNode is one step in the logical/physical hierarchy path for an RSC
// inventory object.
type PathNode struct {
	FID        string `json:"fid"`
	Name       string `json:"name"`
	ObjectType string `json:"objectType"`
}

// NativeResourceGroup represents an Azure resource group surfaced by RSC.
type NativeResourceGroup struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Subscription struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"azureSubscriptionDetails"`
	SLAAssignment sla.Assignment `json:"slaAssignment"`
	LogicalPath   []PathNode     `json:"logicalPath"`
	PhysicalPath  []PathNode     `json:"physicalPath"`
}

// NativeResourceGroups returns all Azure resource groups visible to RSC,
// filtered to the given subscription IDs and (optionally) to those whose name
// contains nameSubstring. The subscription ID list must be non-empty: the
// server rejects an empty filter; the high-level wrapper in
// pkg/polaris/azure/native_account.go expands a nil/empty list to all managed
// subscriptions before calling this method.
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
