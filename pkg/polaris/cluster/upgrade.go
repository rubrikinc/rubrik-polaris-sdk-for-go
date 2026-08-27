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
// registered. Use DownloadPackageAndWait for the race-safe composite.
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

		delay = nextBackoff(delay, upgradePollMaxDelay)
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

// nextBackoff doubles cur, clamping to max. Shared by the upgrade poll loops.
func nextBackoff(cur, max time.Duration) time.Duration {
	if cur >= max {
		return max
	}
	if cur *= 2; cur > max {
		return max
	}
	return cur
}

// transitionPollInitialDelay is the first sleep when polling for a
// post-trigger state transition. Short because the cluster typically picks
// the new operation up within seconds.
var transitionPollInitialDelay = time.Second

// transitionPollMaxDelay caps the exponential backoff between transition
// polls.
var transitionPollMaxDelay = 5 * time.Second

// DownloadPackageAndWait triggers a download of the package at packageURL onto
// the specified cluster and blocks until the package is staged or a terminal
// download failure is observed. The returned CDMInfo is the most recent
// observation regardless of outcome.
//
// Unlike calling DownloadPackage followed by WaitForDownload, this composite
// snapshots the cluster's pre-trigger state and waits for an observed state
// transition before applying completion checks. That avoids a race where the
// cluster was already in READY_FOR_UPGRADE for targetVersion (e.g. a previous
// download attempt), which would otherwise cause WaitForDownload to return
// success before the new operation has registered.
//
// targetVersion is required and must match the version encoded in the package
// URL filename (which the server itself parses for the trigger).
func (a API) DownloadPackageAndWait(ctx context.Context, clusterID uuid.UUID, packageURL, md5checksum, targetVersion string) (gqlcluster.CDMInfo, error) {
	a.log.Print(log.Trace)

	preDetails, err := a.ClusterUpgrade(ctx, clusterID)
	if err != nil {
		return gqlcluster.CDMInfo{}, fmt.Errorf("snapshot pre-trigger cluster upgrade: %s", err)
	}
	var preV2 gqlcluster.RSCUpgradeStatusType
	if preDetails.CDMInfo != nil && preDetails.CDMInfo.UpgradeStatusV2 != nil {
		preV2 = preDetails.CDMInfo.UpgradeStatusV2.RSCClusterUpgradeStatus
	}
	// If the cluster already looks staged for targetVersion, the live V1 state
	// can lead the slower V2 aggregator: releasing on a V1 transition would let
	// WaitForDownload succeed on the stale READY_FOR_UPGRADE@targetVersion. In
	// that case the gate ignores V1, so skip the V1 snapshot entirely. IsStaged
	// is nil-safe.
	preStaged := preDetails.CDMInfo.IsStaged(targetVersion)
	var preV1 gqlcluster.UpgradeState
	if !preStaged {
		if preV1, err = a.UpgradeStatus(ctx, clusterID); err != nil {
			return gqlcluster.CDMInfo{}, fmt.Errorf("snapshot pre-trigger upgrade status: %s", err)
		}
	}
	a.log.Printf(log.Debug, "DownloadPackageAndWait cluster %q baseline: V1=%s V2=%s staged=%t", clusterID, preV1, preV2, preStaged)

	if _, err := a.DownloadPackage(ctx, clusterID, packageURL, md5checksum); err != nil {
		return gqlcluster.CDMInfo{}, err
	}

	if err := a.waitForDownloadTransition(ctx, clusterID, preV1, preV2, targetVersion, preStaged); err != nil {
		return gqlcluster.CDMInfo{}, err
	}

	return a.WaitForDownload(ctx, clusterID, targetVersion)
}

// waitForDownloadTransition polls until the just-triggered operation has
// registered, then returns.
//
// Normally it releases as soon as either V1 (live upgradeStatus) or V2
// (aggregated UpgradeStatusV2) differs from the captured baselines. When the
// cluster was already staged for targetVersion before the trigger (preStaged),
// the V1 signal is ignored: V1 leads the slower V2 aggregator, and releasing on
// V1 alone would let WaitForDownload short-circuit on the stale
// READY_FOR_UPGRADE@targetVersion. In that case the gate releases only once a
// V2 observation is no longer staged for targetVersion.
//
// Transient errors from individual polls are tolerated; only ctx cancellation
// aborts.
func (a API) waitForDownloadTransition(ctx context.Context, clusterID uuid.UUID, preV1 gqlcluster.UpgradeState, preV2 gqlcluster.RSCUpgradeStatusType, targetVersion string, preStaged bool) error {
	delay := transitionPollInitialDelay
	for {
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("wait for cluster %q to register download: %w", clusterID, ctx.Err())
		case <-timer.C:
		}

		// V1 leads the V2 aggregator, so trust it only when the pre-trigger
		// state was not already staged for targetVersion.
		if !preStaged {
			if v1, err := a.UpgradeStatus(ctx, clusterID); err != nil {
				a.log.Printf(log.Warn, "DownloadPackageAndWait V1 poll for cluster %q failed (will retry): %s", clusterID, err)
			} else if v1 != preV1 {
				a.log.Printf(log.Debug, "DownloadPackageAndWait cluster %q V1 transitioned %s -> %s", clusterID, preV1, v1)
				return nil
			}
		}

		if details, err := a.ClusterUpgrade(ctx, clusterID); err != nil {
			a.log.Printf(log.Warn, "DownloadPackageAndWait V2 poll for cluster %q failed (will retry): %s", clusterID, err)
		} else if details.CDMInfo != nil {
			if preStaged {
				// Release only once V2 has left the stale staged state, so
				// WaitForDownload cannot succeed on the pre-trigger observation.
				if !details.CDMInfo.IsStaged(targetVersion) {
					a.log.Printf(log.Debug, "DownloadPackageAndWait cluster %q V2 left stale staged state", clusterID)
					return nil
				}
			} else if v2 := details.CDMInfo.UpgradeStatusV2; v2 != nil && v2.RSCClusterUpgradeStatus != preV2 {
				a.log.Printf(log.Debug, "DownloadPackageAndWait cluster %q V2 transitioned %s -> %s", clusterID, preV2, v2.RSCClusterUpgradeStatus)
				return nil
			}
		}

		delay = nextBackoff(delay, transitionPollMaxDelay)
	}
}

// Upgrade kicks off a fresh upgrade on the specified cluster. upgradeType
// selects between fast (UpgradeTypeFast) and rolling (UpgradeTypeRolling).
// Returns the per-cluster server reply; if the server rejects the job
// (reply.Success false), the returned error wraps reply.Message. Poll
// WaitForUpgrade to block until the cluster is on targetVersion.
//
// The mutation also sets the cluster's stored upgrade-type preference (the
// same value SetUpgradeType configures) to match the chosen upgradeType.
//
// Upgrade only triggers a fresh upgrade (START). Recovery actions (resume a
// failed upgrade, roll back) are not exposed here; callers needing those
// can use gqlcluster.StartUpgrade directly.
func (a API) Upgrade(ctx context.Context, clusterID uuid.UUID, upgradeType gqlcluster.UpgradeType, version string) (gqlcluster.UpgradeJobReply, error) {
	a.log.Print(log.Trace)

	mode := "normal"
	if upgradeType == gqlcluster.UpgradeTypeRolling {
		mode = "rolling"
	}
	reply, err := gqlcluster.StartUpgrade(ctx, a.client.GQL, clusterID, mode, gqlcluster.ActionStart, version, "")
	if err != nil {
		return gqlcluster.UpgradeJobReply{}, fmt.Errorf("failed to start upgrade: %s", err)
	}
	if !reply.Success {
		msg := reply.Message
		if msg == "" {
			msg = "no message returned"
		}
		return reply, fmt.Errorf("cluster %q upgrade rejected: %s", clusterID, msg)
	}
	return reply, nil
}

// WaitForUpgrade polls the cluster's upgrade state until the installed version
// matches targetVersion (success), a terminal upgrade failure is observed
// (returned as a wrapped error), or ctx is cancelled. The returned CDMInfo is
// the most recent observation regardless of outcome. Backoff is exponential,
// capped at upgradePollMaxDelay. Transient errors from individual poll
// requests are logged at Warn and tolerated — only ctx cancellation aborts
// the wait.
//
// An observed ROLLINGBACK status is returned as a wrapped error: once the
// cluster has decided to roll back, the target version will not be reached
// without operator intervention. Use WaitForRollback to wait through the
// rollback itself.
func (a API) WaitForUpgrade(ctx context.Context, clusterID uuid.UUID, targetVersion string) (gqlcluster.CDMInfo, error) {
	a.log.Print(log.Trace)

	delay := upgradePollInitialDelay
	var last gqlcluster.CDMInfo
	for {
		details, err := a.ClusterUpgrade(ctx, clusterID)
		if err != nil {
			a.log.Printf(log.Warn, "WaitForUpgrade poll for cluster %q failed (will retry): %s", clusterID, err)
		} else if details.CDMInfo != nil {
			last = *details.CDMInfo
			if last.Version == targetVersion {
				return last, nil
			}
			if status, failed := upgradeTerminalFailure(last); failed {
				return last, fmt.Errorf("cluster %q upgrade to %q failed: %s", clusterID, targetVersion, status)
			}
			if upgradeStatus(last) == gqlcluster.RSCUpgradeStatusRollingBack {
				return last, fmt.Errorf("cluster %q upgrade to %q is rolling back", clusterID, targetVersion)
			}
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return last, fmt.Errorf("wait for cluster %q upgrade to %q: %w", clusterID, targetVersion, ctx.Err())
		case <-timer.C:
		}

		delay = nextBackoff(delay, upgradePollMaxDelay)
	}
}

// WaitForRollback polls the cluster's upgrade state until a rollback settles.
// When previousVersion is non-empty, success means Version == previousVersion.
// When previousVersion is empty (the caller doesn't know the target — e.g.
// CDMInfo.PreviousVersion was not populated), success means the V2 status has
// left ROLLINGBACK for any non-failure state. ROLLINGBACK_FAILED is returned
// as a wrapped error.
//
// Use this after observing ROLLINGBACK from WaitForUpgrade — WaitForUpgrade
// treats ROLLINGBACK as a terminal failure of the original upgrade attempt,
// so this is the dedicated primitive for waiting through the rollback.
func (a API) WaitForRollback(ctx context.Context, clusterID uuid.UUID, previousVersion string) (gqlcluster.CDMInfo, error) {
	a.log.Print(log.Trace)

	delay := upgradePollInitialDelay
	var last gqlcluster.CDMInfo
	for {
		details, err := a.ClusterUpgrade(ctx, clusterID)
		if err != nil {
			a.log.Printf(log.Warn, "WaitForRollback poll for cluster %q failed (will retry): %s", clusterID, err)
		} else if details.CDMInfo != nil {
			last = *details.CDMInfo
			status := upgradeStatus(last)
			if status == gqlcluster.RSCUpgradeStatusRollingBackFailed {
				target := previousVersion
				if target == "" {
					target = "previous version"
				}
				return last, fmt.Errorf("cluster %q rollback to %q failed", clusterID, target)
			}
			if previousVersion != "" {
				if last.Version == previousVersion {
					return last, nil
				}
			} else if last.UpgradeStatusV2 != nil && status != gqlcluster.RSCUpgradeStatusRollingBack {
				return last, nil
			}
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return last, fmt.Errorf("wait for cluster %q rollback to %q: %w", clusterID, previousVersion, ctx.Err())
		case <-timer.C:
		}

		delay = nextBackoff(delay, upgradePollMaxDelay)
	}
}

// upgradeStatus returns the V2 status if available; otherwise UNKNOWN.
func upgradeStatus(info gqlcluster.CDMInfo) gqlcluster.RSCUpgradeStatusType {
	if info.UpgradeStatusV2 != nil {
		return info.UpgradeStatusV2.RSCClusterUpgradeStatus
	}
	return gqlcluster.RSCUpgradeStatusUnknown
}

// upgradeTerminalFailure reports whether info indicates a terminal upgrade
// failure (UPGRADE_FAILED via V2, or the equivalent V1 cluster-job statuses
// when V2 is absent). The returned string is the status name for use in
// error messages.
func upgradeTerminalFailure(info gqlcluster.CDMInfo) (string, bool) {
	if info.UpgradeStatusV2 != nil {
		if info.UpgradeStatusV2.RSCClusterUpgradeStatus == gqlcluster.RSCUpgradeStatusUpgradeFailed {
			return string(info.UpgradeStatusV2.RSCClusterUpgradeStatus), true
		}
		return "", false
	}
	switch info.ClusterJobStatus {
	case gqlcluster.ClusterJobStatusUpgradeFailed,
		gqlcluster.ClusterJobStatusFailedToInitiateUpgrade:
		return string(info.ClusterJobStatus), true
	}
	return "", false
}
