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
	"time"

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

// upgradePollInitialDelay is the first sleep between WaitForDownload polls.
var upgradePollInitialDelay = 15 * time.Second

// upgradePollMaxDelay caps the exponential backoff between WaitForDownload
// polls.
var upgradePollMaxDelay = 60 * time.Second

// DownloadPackage starts a download of the package at packageURL onto the
// specified cluster. The server parses the version from the package filename,
// which must follow the `rubrik-image-<version>.zip` convention. Returns the
// job ID; poll WaitForDownload to block until the package is staged.
func (a API) DownloadPackage(ctx context.Context, clusterID uuid.UUID, packageURL, md5checksum string) (string, error) {
	a.log.Print(log.Trace)

	jobID, err := gqlcluster.StartDownloadPackage(ctx, a.client.GQL, clusterID, packageURL, md5checksum)
	if err != nil {
		return "", fmt.Errorf("failed to start download: %s", err)
	}
	return jobID, nil
}

// RetryDownloadPackage retries the previously failed download package job
// for the specified cluster. Returns the new job ID.
func (a API) RetryDownloadPackage(ctx context.Context, clusterID uuid.UUID) (string, error) {
	a.log.Print(log.Trace)

	jobID, err := gqlcluster.RetryDownloadPackageJob(ctx, a.client.GQL, clusterID)
	if err != nil {
		return "", fmt.Errorf("failed to retry download package: %s", err)
	}
	return jobID, nil
}

// WaitForDownload polls the cluster's upgrade state until the package is
// staged (success), a terminal download failure is observed (returned as a
// wrapped error), or ctx is cancelled. The returned CDMInfo is the most
// recent observation regardless of outcome. Backoff is exponential, capped at
// upgradePollMaxDelay. Transient errors from individual poll requests are
// logged at Warn and tolerated — only ctx cancellation aborts the wait.
//
// The server runs prechecks synchronously at the tail of the download flow.
// Any state past DOWNLOADING (PRECHECKING, READY_FOR_UPGRADE, etc.) means the
// package is on the cluster, so WaitForDownload returns there rather than
// blocking on prechecks — that's a downstream concern. PRECHECK_FAILED is
// also returned early because it leaves the cluster unable to naturally
// progress to READY_FOR_UPGRADE without operator intervention.
//
// The staged-success path is version-checked: READY_FOR_UPGRADE only counts
// as success (nil error) when its reported target version equals
// targetVersion (see IsStaged). A cluster sitting in READY_FOR_UPGRADE for a
// different version does not short-circuit the wait — it keeps polling until
// the cluster transitions or ctx is cancelled, so callers should always pass
// a ctx with a deadline. The other post-download states (PRECHECKING,
// UPGRADING, etc.) are still returned with a nil error without a version
// check, since reaching them at all implies a package was staged.
//
// Calling DownloadPackage immediately followed by WaitForDownload is
// race-prone: if the cluster was already in READY_FOR_UPGRADE for
// targetVersion (e.g. from a previous successful download), the first poll
// will see that stale state and return before the new operation has even
// registered. Callers wanting trigger-then-wait semantics should capture the
// pre-trigger state and wait for a transition before applying completion
// checks.
func (a API) WaitForDownload(ctx context.Context, clusterID uuid.UUID, targetVersion string) (gqlcluster.CDMInfo, error) {
	a.log.Print(log.Trace)

	delay := upgradePollInitialDelay
	var last gqlcluster.CDMInfo
	for {
		details, err := a.ClusterUpgrade(ctx, clusterID)
		if err != nil {
			a.log.Printf(log.Warn, "WaitForDownload poll for cluster %q failed (will retry): %s", clusterID, err)
		} else if details.CDMInfo != nil {
			last = *details.CDMInfo
			if details.CDMInfo.IsStaged(targetVersion) {
				return last, nil
			}
			if isDownloadTerminalFailure(last) {
				return last, fmt.Errorf("cluster %q download of %q failed: %s",
					clusterID, targetVersion, downloadFailureReason(last))
			}
			// Any state past DOWNLOADING means the package is staged. The
			// version-checked READY_FOR_UPGRADE (handled above) and the
			// pre-progress WAITING_FOR_OPERATION_TO_START are deliberately
			// excluded.
			if v2 := last.UpgradeStatusV2; v2 != nil {
				switch v2.RSCClusterUpgradeStatus {
				case gqlcluster.RSCUpgradeStatusPrechecking,
					gqlcluster.RSCUpgradeStatusPrecheckFailed,
					gqlcluster.RSCUpgradeStatusUpgrading,
					gqlcluster.RSCUpgradeStatusUpgradeFailed,
					gqlcluster.RSCUpgradeStatusRollingBack,
					gqlcluster.RSCUpgradeStatusRollingBackFailed:
					return last, nil
				}
			}
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return last, fmt.Errorf("wait for cluster %q download of %q: %w",
				clusterID, targetVersion, ctx.Err())
		case <-timer.C:
		}

		if delay < upgradePollMaxDelay {
			delay *= 2
			if delay > upgradePollMaxDelay {
				delay = upgradePollMaxDelay
			}
		}
	}
}

func isDownloadTerminalFailure(info gqlcluster.CDMInfo) bool {
	if info.UpgradeStatusV2 != nil {
		return info.UpgradeStatusV2.RSCClusterUpgradeStatus == gqlcluster.RSCUpgradeStatusDownloadFailed
	}
	return info.ClusterJobStatus == gqlcluster.ClusterJobStatusDownloadPackageFailed
}

func downloadFailureReason(info gqlcluster.CDMInfo) string {
	if info.UpgradeStatusV2 != nil {
		return string(info.UpgradeStatusV2.RSCClusterUpgradeStatus)
	}
	return string(info.ClusterJobStatus)
}
