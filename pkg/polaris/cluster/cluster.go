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

// Package cluster provides a high level interface to the cluster part of the RSC
// platform.

package cluster

import (
	"context"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlcluster "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for cluster management.
type API struct {
	client gqlcluster.API
	log    log.Logger
}

// Wrap the RSC client in the cluster API.
func Wrap(client *polaris.Client) API {
	return API{
		client: gqlcluster.Wrap(client.GQL),
		log:    client.GQL.Log(),
	}
}

// SLASourceClusters returns all SLA source clusters.
func (a API) SLASourceClusters(ctx context.Context) ([]gqlcluster.SLADataLocationCluster, error) {
	a.log.Print(log.Trace)

	clusters, err := gqlcluster.SLASourceClusters(ctx, a.client.GQL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list SLA source clusters: %s", err)
	}

	return clusters, nil
}

// SLASourceClusterByName returns the SLA source cluster with the specified name.
func (a API) SLASourceClusterByName(ctx context.Context, name string) (gqlcluster.SLADataLocationCluster, error) {
	a.log.Print(log.Trace)

	var filters []gqlcluster.ClusterFilter
	if name != "" {
		filters = append(filters, gqlcluster.ClusterFilter{
			Field:  "CLUSTER_NAME",
			Values: []string{name},
		})
	}
	clusters, err := gqlcluster.SLASourceClusters(ctx, a.client.GQL, filters)
	if err != nil {
		return gqlcluster.SLADataLocationCluster{}, fmt.Errorf("failed to list SLA source clusters: %s", err)
	}

	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}

	return gqlcluster.SLADataLocationCluster{}, fmt.Errorf("cluster %q %w", name, graphql.ErrNotFound)
}
