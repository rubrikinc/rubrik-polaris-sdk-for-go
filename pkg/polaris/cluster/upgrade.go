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

package cluster

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlcluster "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// listClusterUpgradesPageSize is the per-request page size used when
// iterating clusterWithUpgradesInfo.
const listClusterUpgradesPageSize = 50

// ListClusterUpgrades returns the upgrade information for every cluster
// matching filter, paginating through the underlying connection.
func (a API) ListClusterUpgrades(ctx context.Context, filter *gqlcluster.CDMInfoFilter, sortBy gqlcluster.UpgradeInfoSortBy, sortOrder core.SortOrder) ([]gqlcluster.UpgradeDetails, error) {
	a.log.Print(log.Trace)

	pageSize := listClusterUpgradesPageSize
	page := core.Pagination{First: &pageSize}
	if sortOrder != "" {
		page.SortOrder = &sortOrder
	}

	var details []gqlcluster.UpgradeDetails
	for {
		result, err := gqlcluster.ClusterWithUpgradesInfo(ctx, a.client.GQL, filter, sortBy, page)
		if err != nil {
			return nil, fmt.Errorf("failed to list cluster upgrades: %s", err)
		}
		details = append(details, result.Details...)
		if !result.PageInfo.HasNextPage || result.PageInfo.EndCursor == "" {
			return details, nil
		}
		if page.After != nil && *page.After == result.PageInfo.EndCursor {
			return details, nil
		}
		cursor := result.PageInfo.EndCursor
		page.After = &cursor
	}
}

// ClusterUpgrade returns the upgrade information for a single cluster.
// Returns graphql.ErrNotFound when no cluster with that UUID is registered.
func (a API) ClusterUpgrade(ctx context.Context, clusterID uuid.UUID) (gqlcluster.UpgradeDetails, error) {
	a.log.Print(log.Trace)

	pageSize := 1
	page := core.Pagination{First: &pageSize}
	filter := &gqlcluster.CDMInfoFilter{ID: []uuid.UUID{clusterID}}
	result, err := gqlcluster.ClusterWithUpgradesInfo(ctx, a.client.GQL, filter, "", page)
	if err != nil {
		return gqlcluster.UpgradeDetails{}, fmt.Errorf("failed to get cluster upgrade info: %s", err)
	}
	if len(result.Details) == 0 || result.Details[0].ID != clusterID {
		return gqlcluster.UpgradeDetails{}, fmt.Errorf("cluster %q %w", clusterID, graphql.ErrNotFound)
	}
	return result.Details[0], nil
}

// SelfServeRollingUpgrade returns whether the account-level self-serve
// rolling upgrade setting is enabled.
func (a API) SelfServeRollingUpgrade(ctx context.Context) (bool, error) {
	a.log.Print(log.Trace)

	enabled, err := gqlcluster.SelfServeRollingUpgrade(ctx, a.client.GQL)
	if err != nil {
		return false, fmt.Errorf("failed to read self-serve rolling upgrade setting: %s", err)
	}
	return enabled, nil
}

// SetSelfServeRollingUpgrade sets the account-level self-serve rolling upgrade
// setting to the specified value.
func (a API) SetSelfServeRollingUpgrade(ctx context.Context, enabled bool) error {
	a.log.Print(log.Trace)

	if err := gqlcluster.SetSelfServeRollingUpgrade(ctx, a.client.GQL, enabled); err != nil {
		return fmt.Errorf("failed to set self-serve rolling upgrade: %s", err)
	}
	return nil
}

// SetUpgradeType sets the upgrade type (fast or rolling) for the specified
// cluster. The returned reply carries the server-side status code and message;
// callers should inspect Code for operational outcomes.
func (a API) SetUpgradeType(ctx context.Context, clusterID uuid.UUID, upgradeType gqlcluster.UpgradeType) (gqlcluster.SetUpgradeTypeReply, error) {
	a.log.Print(log.Trace)

	reply, err := gqlcluster.SetUpgradeType(ctx, a.client.GQL, clusterID, upgradeType)
	if err != nil {
		return gqlcluster.SetUpgradeTypeReply{}, fmt.Errorf("failed to set cluster upgrade type: %s", err)
	}
	return reply, nil
}

// UpgradeStatus returns the cluster's real-time upgrade state via a live read
// from the upgrade service. Prefer this over ClusterUpgrade when authoritative
// state is required.
func (a API) UpgradeStatus(ctx context.Context, clusterID uuid.UUID) (gqlcluster.UpgradeState, error) {
	a.log.Print(log.Trace)

	state, err := gqlcluster.UpgradeStatus(ctx, a.client.GQL, clusterID)
	if err != nil {
		return "", fmt.Errorf("failed to read upgrade status: %s", err)
	}
	return state, nil
}
