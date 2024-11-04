// Copyright 2024 Rubrik, Inc.
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

package exocompute

import (
	"context"
	"encoding/json"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ClusterConnectionParams holds the parameters for an exocompute cluster
// connection info operation.
type ClusterConnectionParams[R ClusterConnectionResult] interface {
	InfoQuery() (string, any, R)
}

// ClusterConnectionResult holds the result of an exocompute cluster connection
// info operation.
type ClusterConnectionResult interface {
	AWSClusterConnectionResult | AzureClusterConnectionResult
}

// ClusterConnection returns information about the connection to the exocompute
// cluster.
func ClusterConnection[P ClusterConnectionParams[R], R ClusterConnectionResult](ctx context.Context, gql *graphql.Client, params P) (R, error) {
	gql.Log().Print(log.Trace)

	query, queryParams, defValue := params.InfoQuery()
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return defValue, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return defValue, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// ConnectClusterParams holds the parameters for an exocompute cluster connect
// operation.
type ConnectClusterParams[R ConnectClusterResult] interface {
	ConnectQuery() (string, any, R)
}

// ConnectClusterResult holds the result of an exocompute cluster connect
// operation.
type ConnectClusterResult interface {
	ConnectAWSClusterResult | ConnectAzureClusterResult
}

// ConnectCluster connects the exocompute cluster to RSC.
func ConnectCluster[P ConnectClusterParams[R], R ConnectClusterResult](ctx context.Context, gql *graphql.Client, params P) (R, error) {
	gql.Log().Print(log.Trace)

	query, queryParams, defValue := params.ConnectQuery()
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return defValue, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return defValue, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// DisconnectClusterParams holds the parameters for an exocompute cluster
// disconnect operation.
type DisconnectClusterParams[R DisconnectClusterResult] interface {
	DisconnectQuery() (string, any, R)
}

// DisconnectClusterResult holds the result of an exocompute cluster disconnect
// operation.
type DisconnectClusterResult interface {
	DisconnectAWSClusterResult | DisconnectAzureClusterResult
}

// DisconnectCluster disconnects the exocompute cluster from RSC.
func DisconnectCluster[P DisconnectClusterParams[R], R DisconnectClusterResult](ctx context.Context, gql *graphql.Client, params P) error {
	gql.Log().Print(log.Trace)

	query, queryParams, _ := params.DisconnectQuery()
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}
