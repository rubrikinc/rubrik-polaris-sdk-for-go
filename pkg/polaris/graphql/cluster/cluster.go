//go:generate go run ../queries_gen.go cluster

// Copyright 2025 Rubrik, Inc.
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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the RSC Cluster API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the GraphQL client in the Cluster API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// ClusterFilter represents a filter for SLA source clusters.
type ClusterFilter struct {
	Field  string   `json:"field"`
	Values []string `json:"texts"`
}

// SLADataLocationCluster represents a cluster in the SLA data location.
type SLADataLocationCluster struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	ClusterInfo struct {
		IsConnected bool `json:"isConnected"`
	} `json:"clusterInfo"`
}

// SLASourceClusters returns all SLA source clusters matching the specified filters.
func SLASourceClusters(ctx context.Context, gql *graphql.Client, filters []ClusterFilter) ([]SLADataLocationCluster, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var clusters []SLADataLocationCluster
	for {
		query := slaSourceClustersQuery
		buf, err := gql.Request(ctx, query, struct {
			First               int             `json:"first,omitempty"`
			After               string          `json:"after,omitempty"`
			SortBy              string          `json:"sortBy,omitempty"`
			SortOrder           core.SortOrder  `json:"sortOrder,omitempty"`
			Filter              []ClusterFilter `json:"filter"`
			IsArchivalSelected  bool            `json:"isArchivalSelected"`
			SelectedReplication string          `json:"selectedReplication,omitempty"`
			SLAID               *string         `json:"slaId,omitempty"`
		}{First: 50, After: cursor, SortBy: "CLUSTER_NAME", SortOrder: core.SortOrderAsc, Filter: filters, IsArchivalSelected: false, SelectedReplication: "ONE_TO_ONE", SLAID: nil})
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Edges []struct {
						Cursor string `json:"cursor"`
						Node   struct {
							DisableReasons      []string               `json:"disableReasons"`
							HasProtectedObjects bool                   `json:"hasProtectedObjects"`
							Cluster             SLADataLocationCluster `json:"cluster"`
						} `json:"node"`
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
			clusters = append(clusters, edge.Node.Cluster)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return clusters, nil
}
