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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// UpgradeInfoPage is one page of results from clusterWithUpgradesInfo.
type UpgradeInfoPage struct {
	Details  []UpgradeDetails
	PageInfo core.PageInfo
	Count    int
}

// ClusterWithUpgradesInfo returns one page of cluster upgrade information,
// applying the optional filter and sort and using the supplied pagination.
func ClusterWithUpgradesInfo(ctx context.Context, gql *graphql.Client, filter *CDMInfoFilter, sortBy UpgradeInfoSortBy, page core.Pagination) (UpgradeInfoPage, error) {
	gql.Log().Print(log.Trace)

	if filter == nil {
		filter = &CDMInfoFilter{}
	}
	query := clusterWithUpgradesInfoQuery
	buf, err := gql.Request(ctx, query, struct {
		First                 *int               `json:"first,omitempty"`
		After                 *string            `json:"after,omitempty"`
		Last                  *int               `json:"last,omitempty"`
		Before                *string            `json:"before,omitempty"`
		SortOrder             *core.SortOrder    `json:"sortOrder,omitempty"`
		SortBy                UpgradeInfoSortBy  `json:"sortBy,omitempty"`
		ClusterLocation       []string           `json:"clusterLocation,omitempty"`
		ConnectionState       []Status           `json:"connectionState,omitempty"`
		DownloadedVersion     []string           `json:"downloadedVersion,omitempty"`
		EOSStatus             []ClusterEOSStatus `json:"eosStatus,omitempty"`
		ID                    []uuid.UUID        `json:"id,omitempty"`
		InstalledVersion      []string           `json:"installedVersion,omitempty"`
		MinSoftwareVersion    string             `json:"minSoftwareVersion,omitempty"`
		Name                  []string           `json:"name,omitempty"`
		PrechecksStatus       []PrechecksStatus  `json:"prechecksStatus,omitempty"`
		ProductType           []Product          `json:"productType,omitempty"`
		RegistrationTimeGT    *time.Time         `json:"registrationTimeGt,omitzero"`
		RegistrationTimeLT    *time.Time         `json:"registrationTimeLt,omitzero"`
		Type                  []ProductType      `json:"type,omitempty"`
		UpgradeJobStatus      []ClusterJobStatus `json:"upgradeJobStatus,omitempty"`
		UpgradeScheduled      *bool              `json:"upgradeScheduled,omitempty"`
		UpgradeStatusCategory []string           `json:"upgradeStatusCategory,omitempty"`
		VersionStatus         []VersionStatus    `json:"versionStatus,omitempty"`
	}{
		First:                 page.First,
		After:                 page.After,
		Last:                  page.Last,
		Before:                page.Before,
		SortOrder:             page.SortOrder,
		SortBy:                sortBy,
		ClusterLocation:       filter.ClusterLocation,
		ConnectionState:       filter.ConnectionState,
		DownloadedVersion:     filter.DownloadedVersion,
		EOSStatus:             filter.EOSStatus,
		ID:                    filter.ID,
		InstalledVersion:      filter.InstalledVersion,
		MinSoftwareVersion:    filter.MinSoftwareVersion,
		Name:                  filter.Name,
		PrechecksStatus:       filter.PrechecksStatus,
		ProductType:           filter.ProductType,
		RegistrationTimeGT:    filter.RegistrationTimeGT,
		RegistrationTimeLT:    filter.RegistrationTimeLT,
		Type:                  filter.Type,
		UpgradeJobStatus:      filter.UpgradeJobStatus,
		UpgradeScheduled:      filter.UpgradeScheduled,
		UpgradeStatusCategory: filter.UpgradeStatusCategory,
		VersionStatus:         filter.VersionStatus,
	})
	if err != nil {
		return UpgradeInfoPage{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Edges []struct {
					Node UpgradeDetails `json:"node"`
				} `json:"edges"`
				PageInfo core.PageInfo `json:"pageInfo"`
				Count    int           `json:"count"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return UpgradeInfoPage{}, graphql.UnmarshalError(query, err)
	}

	details := make([]UpgradeDetails, 0, len(payload.Data.Result.Edges))
	for _, edge := range payload.Data.Result.Edges {
		details = append(details, edge.Node)
	}
	return UpgradeInfoPage{
		Details:  details,
		PageInfo: payload.Data.Result.PageInfo,
		Count:    payload.Data.Result.Count,
	}, nil
}

// ReleaseDetail describes one available CDM release as reported by
// cdmReleaseDetailsForClusterFromSupportPortal. The fields used to drive a
// download are Name (version), URL, MD5Sum, Size.
type ReleaseDetail struct {
	Name        string `json:"name"`
	Recommended bool   `json:"isRecommended"`
	Upgradable  bool   `json:"isUpgradable"`
	MD5Sum      string `json:"md5Sum"`
	Size        int64  `json:"size"`
	URL         string `json:"tarDownloadLink"`
}

// ListUpgradesOptions controls the filter knobs of the underlying query.
// All boolean fields default to false; callers must opt in to each filter.
// SortOrder is the only truly optional field — leave it empty to let the
// server decide.
type ListUpgradesOptions struct {
	FilterVersion     string
	FetchLinks        bool
	FilterUpgradeable bool
	ShouldShowAll     bool
	FilterAfterSource bool
	SortOrder         core.SortOrder
}

// ListUpgrades returns the available CDM releases for the supplied clusters,
// optionally filtered by version (server-defined match semantics). RSC fetches
// the underlying data from the Rubrik support portal on the SDK's behalf.
func ListUpgrades(ctx context.Context, gql *graphql.Client, clusters []uuid.UUID, opts ListUpgradesOptions) ([]ReleaseDetail, error) {
	gql.Log().Print(log.Trace)

	if len(clusters) == 0 {
		return nil, fmt.Errorf("at least one cluster UUID is required")
	}

	query := cdmReleaseDetailsForClusterFromSupportPortalQuery
	vars := struct {
		ListClusterUUID   []uuid.UUID     `json:"listClusterUuid"`
		FilterVersion     string          `json:"filterVersion"`
		FetchLinks        bool            `json:"fetchLinks"`
		FilterUpgradeable bool            `json:"filterUpgradeable"`
		ShouldShowAll     bool            `json:"shouldShowAll"`
		FilterAfterSource bool            `json:"filterAfterSource"`
		SortOrder         *core.SortOrder `json:"sortOrder,omitempty"`
	}{
		ListClusterUUID:   clusters,
		FilterVersion:     opts.FilterVersion,
		FetchLinks:        opts.FetchLinks,
		FilterUpgradeable: opts.FilterUpgradeable,
		ShouldShowAll:     opts.ShouldShowAll,
		FilterAfterSource: opts.FilterAfterSource,
	}
	if opts.SortOrder != "" {
		vars.SortOrder = &opts.SortOrder
	}
	buf, err := gql.Request(ctx, query, vars)
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ReleaseDetails []ReleaseDetail `json:"releaseDetails"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}
	return payload.Data.Result.ReleaseDetails, nil
}

// MultiHopUpgradePath returns the ordered sequence of CDM versions required
// to upgrade the cluster from sourceVersion to targetVersion (both inclusive).
// If sourceVersion is empty, the server uses the cluster's currently installed
// version. When fullVersionName is true, each hop is returned as the full
// release name including patch and build number.
func MultiHopUpgradePath(ctx context.Context, gql *graphql.Client, clusterID uuid.UUID, sourceVersion, targetVersion string, fullVersionName bool) ([]string, error) {
	gql.Log().Print(log.Trace)

	if targetVersion == "" {
		return nil, fmt.Errorf("target version is required")
	}

	query := multiHopUpgradePathQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID                  uuid.UUID `json:"clusterUuid"`
		SourceVersion                string    `json:"sourceVersion,omitempty"`
		TargetVersion                string    `json:"targetVersion"`
		ShouldIncludeFullVersionName bool      `json:"shouldIncludeFullVersionName"`
	}{
		ClusterUUID:                  clusterID,
		SourceVersion:                sourceVersion,
		TargetVersion:                targetVersion,
		ShouldIncludeFullVersionName: fullVersionName,
	})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				VersionPath []string `json:"versionPath"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}
	return payload.Data.Result.VersionPath, nil
}

// SelfServeRollingUpgrade returns whether self-serve rolling upgrade is
// enabled for the account.
func SelfServeRollingUpgrade(ctx context.Context, gql *graphql.Client) (bool, error) {
	gql.Log().Print(log.Trace)

	query := selfServeRollingUpgradeQuery
	buf, err := gql.Request(ctx, query, struct{}{})
	if err != nil {
		return false, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Enabled bool `json:"enabled"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, graphql.UnmarshalError(query, err)
	}
	return payload.Data.Result.Enabled, nil
}

// SetSelfServeRollingUpgrade sets the account-level self-serve rolling upgrade
// setting to the specified value.
func SetSelfServeRollingUpgrade(ctx context.Context, gql *graphql.Client, enabled bool) error {
	gql.Log().Print(log.Trace)

	query := setSelfServeRollingUpgradeQuery
	buf, err := gql.Request(ctx, query, struct {
		Enabled bool `json:"enabled"`
	}{Enabled: enabled})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Enabled bool `json:"enabled"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result.Enabled != enabled {
		return graphql.ResponseError(query, fmt.Errorf("self-serve rolling upgrade not set: requested %t, got %t", enabled, payload.Data.Result.Enabled))
	}
	return nil
}

// UpgradeType represents the per-cluster upgrade type preference.
type UpgradeType string

const (
	UpgradeTypeFast    UpgradeType = "FAST"
	UpgradeTypeRolling UpgradeType = "ROLLING"
)

// ActionType represents the upgrade action.
type ActionType string

const (
	ActionResume   ActionType = "RESUME"
	ActionRollback ActionType = "ROLLBACK"
	ActionStart    ActionType = "START"
)

// UpgradeJobReply is the per-cluster reply from startUpgradeBatchJob (and
// scheduleUpgradeBatchJob). Success=false signals the server rejected the
// job; Message carries the rejection reason.
type UpgradeJobReply struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// StartUpgrade kicks off an immediate upgrade for the specified cluster.
// mode is the upgrade mode (e.g. "normal", "rolling"); action is one of the
// ActionType constants. contextTag is optional — pass an empty string for the
// server default.
func StartUpgrade(ctx context.Context, gql *graphql.Client, clusterID uuid.UUID, mode string, action ActionType, version, contextTag string) (UpgradeJobReply, error) {
	gql.Log().Print(log.Trace)

	query := startUpgradeBatchJobQuery
	vars := struct {
		ListClusterUUID []uuid.UUID `json:"listClusterUuid"`
		Mode            string      `json:"mode"`
		Action          ActionType  `json:"action"`
		Version         string      `json:"version"`
		ContextTag      *string     `json:"contextTag,omitempty"`
	}{
		ListClusterUUID: []uuid.UUID{clusterID},
		Mode:            mode,
		Action:          action,
		Version:         version,
	}
	if contextTag != "" {
		vars.ContextTag = &contextTag
	}
	buf, err := gql.Request(ctx, query, vars)
	if err != nil {
		return UpgradeJobReply{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []struct {
				UpgradeJobReply UpgradeJobReply `json:"upgradeJobReply"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return UpgradeJobReply{}, graphql.UnmarshalError(query, err)
	}
	if len(payload.Data.Result) != 1 {
		return UpgradeJobReply{}, graphql.ResponseError(query, fmt.Errorf("expected 1 reply, got %d", len(payload.Data.Result)))
	}
	return payload.Data.Result[0].UpgradeJobReply, nil
}

// SetUpgradeTypeReply is returned by setUpgradeType. The JSON tag for
// Exceptions preserves the schema-level typo ("excepshuns").
type SetUpgradeTypeReply struct {
	Code       string `json:"code"`
	Exceptions string `json:"excepshuns"`
	Message    string `json:"message"`
}

// SetUpgradeType sets the upgrade type (fast or rolling) for the specified
// cluster.
func SetUpgradeType(ctx context.Context, gql *graphql.Client, clusterID uuid.UUID, upgradeType UpgradeType) (SetUpgradeTypeReply, error) {
	gql.Log().Print(log.Trace)

	query := setUpgradeTypeQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID uuid.UUID   `json:"clusterUuid"`
		UpgradeType UpgradeType `json:"upgradeType"`
	}{
		ClusterUUID: clusterID,
		UpgradeType: upgradeType,
	})
	if err != nil {
		return SetUpgradeTypeReply{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result SetUpgradeTypeReply `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return SetUpgradeTypeReply{}, graphql.UnmarshalError(query, err)
	}
	return payload.Data.Result, nil
}

// UpgradeState is the cluster's current upgrade-state-machine state, as
// reported by the upgradeStatus query. The set below covers every value the
// upgrade service is known to emit; new entries may be added by the server
// before the SDK is updated, so callers comparing against an unknown value
// should fall through gracefully.
type UpgradeState string

const (
	UpgradeStateIdle              UpgradeState = "IDLE"
	UpgradeStateAcquiring         UpgradeState = "ACQUIRING"
	UpgradeStateDeploying         UpgradeState = "DEPLOYING"
	UpgradeStatePrechecking       UpgradeState = "PRECHECKING"
	UpgradeStateUpgrading         UpgradeState = "UPGRADING"
	UpgradeStateError             UpgradeState = "ERROR"
	UpgradeStateCopying           UpgradeState = "COPYING"
	UpgradeStateVerifying         UpgradeState = "VERIFYING"
	UpgradeStateUntaring          UpgradeState = "UNTARING"
	UpgradeStatePreparing         UpgradeState = "PREPARING"
	UpgradeStateRestarting        UpgradeState = "RESTARTING"
	UpgradeStateImaging           UpgradeState = "IMAGING"
	UpgradeStateConfiguring       UpgradeState = "CONFIGURING"
	UpgradeStateMigrating         UpgradeState = "MIGRATING"
	UpgradeStateRollingBack       UpgradeState = "ROLLING_BACK"
	UpgradeStateStaged            UpgradeState = "STAGED"
	UpgradeStateRUPrecheckLock    UpgradeState = "RU_PRECHECK_LOCK"
	UpgradeStateRUCRDBUpgrade     UpgradeState = "RU_CRDB_UPGRADE"
	UpgradeStateRUMetadataUpgrade UpgradeState = "RU_METADATA_UPGRADE"
	UpgradeStateRollingUpgrade    UpgradeState = "ROLLING_UPGRADE"
	UpgradeStateRUClusterWrapup   UpgradeState = "RU_CLUSTER_WRAPUP"
	UpgradeStateRUIdle            UpgradeState = "RU_IDLE"
	UpgradeStateRUPrecheck        UpgradeState = "RU_PRECHECK"
	UpgradeStateRUPreparing       UpgradeState = "RU_PREPARING"
	UpgradeStateRUConfiguring     UpgradeState = "RU_CONFIGURING"
	UpgradeStateRUMigrating       UpgradeState = "RU_MIGRATING"
	UpgradeStateRURestarting      UpgradeState = "RU_RESTARTING"
	UpgradeStateRUDone            UpgradeState = "RU_DONE"
	UpgradeStateUnknown           UpgradeState = "UNKNOWN"
)

// UpgradeStatus returns the cluster's real-time upgrade state via a live read
// from the upgrade service, bypassing the aggregated UpgradeStatusV2 view in
// ClusterUpgrade which can lag by tracker-poll intervals.
func UpgradeStatus(ctx context.Context, gql *graphql.Client, clusterID uuid.UUID) (UpgradeState, error) {
	gql.Log().Print(log.Trace)

	query := upgradeStatusQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID uuid.UUID `json:"clusterUuid"`
	}{ClusterUUID: clusterID})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				CurrentStateName UpgradeState `json:"currentStateName"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}
	return payload.Data.Result.CurrentStateName, nil
}

// StartDownloadPackage initiates a download of the package at packageURL onto
// the specified cluster. The server parses the version from the package
// filename, which must follow the `rubrik-image-<version>.zip` convention.
// Returns the per-cluster job ID; poll the cluster's upgrade status to track
// progress.
func StartDownloadPackage(ctx context.Context, gql *graphql.Client, clusterID uuid.UUID, packageURL, md5checksum string) (string, error) {
	gql.Log().Print(log.Trace)

	query := startDownloadPackageBatchJobQuery
	buf, err := gql.Request(ctx, query, struct {
		ListClusterUUID []uuid.UUID `json:"listClusterUuid"`
		PackageURL      string      `json:"packageUrl"`
		Md5Checksum     string      `json:"md5checksum"`
	}{
		ListClusterUUID: []uuid.UUID{clusterID},
		PackageURL:      packageURL,
		Md5Checksum:     md5checksum,
	})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []struct {
				JobID string `json:"jobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}
	if len(payload.Data.Result) != 1 {
		return "", graphql.ResponseError(query, fmt.Errorf("expected 1 reply, got %d", len(payload.Data.Result)))
	}
	return payload.Data.Result[0].JobID, nil
}

// RetryDownloadPackageJob retries the previously failed download package job
// for the specified cluster. Returns the new job ID.
func RetryDownloadPackageJob(ctx context.Context, gql *graphql.Client, clusterID uuid.UUID) (string, error) {
	gql.Log().Print(log.Trace)

	query := retryDownloadPackageJobQuery
	buf, err := gql.Request(ctx, query, struct {
		ClusterUUID uuid.UUID `json:"clusterUuid"`
	}{ClusterUUID: clusterID})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				JobID string `json:"jobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}
	return payload.Data.Result.JobID, nil
}

// UpgradeInfoSortBy represents the sort field for cluster upgrade info queries.
type UpgradeInfoSortBy string

const (
	UpgradeInfoSortByClusterJobStatus  UpgradeInfoSortBy = "ClusterJobStatus"
	UpgradeInfoSortByClusterLocation   UpgradeInfoSortBy = "ClusterLocation"
	UpgradeInfoSortByClusterName       UpgradeInfoSortBy = "ClusterName"
	UpgradeInfoSortByClusterType       UpgradeInfoSortBy = "ClusterType"
	UpgradeInfoSortByDownloadedVersion UpgradeInfoSortBy = "DownloadedVersion"
	UpgradeInfoSortByInstalledVersion  UpgradeInfoSortBy = "InstalledVersion"
	UpgradeInfoSortByRegisteredAt      UpgradeInfoSortBy = "RegisteredAt"
	UpgradeInfoSortByUpgradeType       UpgradeInfoSortBy = "UpgradeType"
	UpgradeInfoSortByVersionStatus     UpgradeInfoSortBy = "VersionStatus"
)

// ClusterJobStatus represents the cluster upgrade job status.
type ClusterJobStatus string

const (
	ClusterJobStatusDownloadPackageFailed   ClusterJobStatus = "DownloadPackageFailed"
	ClusterJobStatusDownloadingPackage      ClusterJobStatus = "DownloadingPackage"
	ClusterJobStatusFailedToInitiateUpgrade ClusterJobStatus = "FailedToInitiateUpgrade"
	ClusterJobStatusPreCheckFailureError    ClusterJobStatus = "PreCheckFailureError"
	ClusterJobStatusPreCheckFailureWarning  ClusterJobStatus = "PreCheckFailureWarning"
	ClusterJobStatusReadyForDownload        ClusterJobStatus = "ReadyForDownload"
	ClusterJobStatusReadyForUpgrade         ClusterJobStatus = "ReadyForUpgrade"
	ClusterJobStatusResumingUpgrade         ClusterJobStatus = "ResumingUpgrade"
	ClusterJobStatusRollbackFailed          ClusterJobStatus = "RollbackFailed"
	ClusterJobStatusRollingBackUpgrade      ClusterJobStatus = "RollingBackUpgrade"
	ClusterJobStatusUnknown                 ClusterJobStatus = "Unknown"
	ClusterJobStatusUpToDate                ClusterJobStatus = "UpToDate"
	ClusterJobStatusUpgradeFailed           ClusterJobStatus = "UpgradeFailed"
	ClusterJobStatusUpgrading               ClusterJobStatus = "Upgrading"
)

// ClusterEOSStatus represents the end-of-support status of a cluster.
type ClusterEOSStatus string

const (
	ClusterEOSStatusPlanUpgrade ClusterEOSStatus = "EOS_STATUS_PLAN_UPGRADE"
	ClusterEOSStatusSupported   ClusterEOSStatus = "EOS_STATUS_SUPPORTED"
	ClusterEOSStatusUnknown     ClusterEOSStatus = "EOS_STATUS_UNKNOWN"
	ClusterEOSStatusUnsupported ClusterEOSStatus = "EOS_STATUS_UNSUPPORTED"
)

// PrechecksStatus represents the precheck status of a cluster.
type PrechecksStatus string

const (
	PrechecksStatusFailureError   PrechecksStatus = "PrechecksFailureError"
	PrechecksStatusFailureWarning PrechecksStatus = "PrechecksFailureWarning"
	PrechecksStatusRunning        PrechecksStatus = "PrechecksRunning"
	PrechecksStatusSuccess        PrechecksStatus = "PrechecksSuccess"
	PrechecksStatusUnknown        PrechecksStatus = "Unknown"
)

// VersionStatus represents the version status of a cluster.
type VersionStatus string

const (
	VersionStatusStable             VersionStatus = "STABLE"
	VersionStatusUnknown            VersionStatus = "UNKNOWN"
	VersionStatusUpgradeRecommended VersionStatus = "UPGRADE_RECOMMENDED"
)

// CDMInfoFilter is the filter input for clusterWithUpgradesInfo.
type CDMInfoFilter struct {
	ClusterLocation       []string           `json:"clusterLocation,omitempty"`
	ConnectionState       []Status           `json:"connectionState,omitempty"`
	DownloadedVersion     []string           `json:"downloadedVersion,omitempty"`
	EOSStatus             []ClusterEOSStatus `json:"eosStatus,omitempty"`
	ID                    []uuid.UUID        `json:"id,omitempty"`
	InstalledVersion      []string           `json:"installedVersion,omitempty"`
	MinSoftwareVersion    string             `json:"minSoftwareVersion,omitempty"`
	Name                  []string           `json:"name,omitempty"`
	PrechecksStatus       []PrechecksStatus  `json:"prechecksStatus,omitempty"`
	ProductType           []Product          `json:"productType,omitempty"`
	RegistrationTimeGT    *time.Time         `json:"registrationTime_gt,omitzero"`
	RegistrationTimeLT    *time.Time         `json:"registrationTime_lt,omitzero"`
	Type                  []ProductType      `json:"type,omitempty"`
	UpgradeJobStatus      []ClusterJobStatus `json:"upgradeJobStatus,omitempty"`
	UpgradeScheduled      *bool              `json:"upgradeScheduled,omitempty"`
	UpgradeStatusCategory []string           `json:"upgradeStatusCategory,omitempty"`
	VersionStatus         []VersionStatus    `json:"versionStatus,omitempty"`
}

// RSCUpgradeStatusType is the V2 enum that supersedes ClusterJobStatus for
// reporting cluster upgrade progress. The V1 ClusterJobStatus field can lag
// (e.g. on stub clusters), so the V2 status is preferred when available.
type RSCUpgradeStatusType string

const (
	RSCUpgradeStatusCDMOnlyOperation         RSCUpgradeStatusType = "CDM_ONLY_OPERATION"
	RSCUpgradeStatusDisconnected             RSCUpgradeStatusType = "DISCONNECTED"
	RSCUpgradeStatusDownloading              RSCUpgradeStatusType = "DOWNLOADING"
	RSCUpgradeStatusDownloadFailed           RSCUpgradeStatusType = "DOWNLOAD_FAILED"
	RSCUpgradeStatusInitializing             RSCUpgradeStatusType = "INITIALIZING"
	RSCUpgradeStatusPrechecking              RSCUpgradeStatusType = "PRECHECKING"
	RSCUpgradeStatusPrecheckFailed           RSCUpgradeStatusType = "PRECHECK_FAILED"
	RSCUpgradeStatusReadyForDownload         RSCUpgradeStatusType = "READY_FOR_DOWNLOAD"
	RSCUpgradeStatusReadyForUpgrade          RSCUpgradeStatusType = "READY_FOR_UPGRADE"
	RSCUpgradeStatusRollingBack              RSCUpgradeStatusType = "ROLLINGBACK"
	RSCUpgradeStatusRollingBackFailed        RSCUpgradeStatusType = "ROLLINGBACK_FAILED"
	RSCUpgradeStatusUnknown                  RSCUpgradeStatusType = "UNKNOWN"
	RSCUpgradeStatusUpgradeFailed            RSCUpgradeStatusType = "UPGRADE_FAILED"
	RSCUpgradeStatusUpgrading                RSCUpgradeStatusType = "UPGRADING"
	RSCUpgradeStatusWaitingForOperationStart RSCUpgradeStatusType = "WAITING_FOR_OPERATION_TO_START"
)

// UIStatusAttributes is the V2 detail payload for the cluster's upgrade
// status. Fields are nullable in the schema; empty values mean "not
// applicable for the current state".
type UIStatusAttributes struct {
	SourceVersion string  `json:"sourceVersion,omitempty"`
	TargetVersion string  `json:"targetVersion,omitempty"`
	Progress      float64 `json:"progress"`
	ErrorMsg      string  `json:"errorMsg,omitempty"`
	UpgradeMode   string  `json:"upgradeMode,omitempty"`
}

// UpgradeStatusV2 is the authoritative upgrade-status payload (preferred
// over the V1 ClusterJobStatus + DownloadedVersion fields when present).
type UpgradeStatusV2 struct {
	RSCClusterUpgradeStatus RSCUpgradeStatusType `json:"rscClusterUpgradeStatus"`
	UIStatus                string               `json:"uiStatus"`
	UIStatusAttributes      UIStatusAttributes   `json:"uiStatusAttributes"`
}

// CDMClusterStatusInfo describes the upgrade tasks state.
type CDMClusterStatusInfo struct {
	CompletedNodes string `json:"completedNodes,omitempty"`
	CurrentNode    string `json:"currentNode,omitempty"`
}

// CDMClusterStatus describes the cluster upgrade status.
type CDMClusterStatus struct {
	Message    string                `json:"message,omitempty"`
	Status     string                `json:"status,omitempty"`
	StatusInfo *CDMClusterStatusInfo `json:"statusInfo,omitzero"`
}

// UpgradeRecommendationInfo describes recommended upgrade versions.
type UpgradeRecommendationInfo struct {
	Recommendation            string   `json:"recommendation"`
	NextReleaseRecommendation string   `json:"nextReleaseRecommendation"`
	Upgradability             []string `json:"upgradability"`
}

// UpgradeDuration describes the duration of the last successful upgrade.
type UpgradeDuration struct {
	ClusterID              uuid.UUID `json:"clusterUuid"`
	FastUpgradeDuration    int64     `json:"fastUpgradeDuration"`
	RollingUpgradeDuration int64     `json:"rollingUpgradeDuration"`
}

// CDMInfo describes the upgrade state of a CDM cluster.
type CDMInfo struct {
	ClusterID                 uuid.UUID                  `json:"clusterUuid"`
	Version                   string                     `json:"version"`
	VersionStatus             VersionStatus              `json:"versionStatus,omitempty"`
	ClusterJobStatus          ClusterJobStatus           `json:"clusterJobStatus,omitempty"`
	ClusterStatus             *CDMClusterStatus          `json:"clusterStatus,omitzero"`
	CurrentStateProgress      float64                    `json:"currentStateProgress"`
	DownloadedVersion         string                     `json:"downloadedVersion,omitempty"`
	FastUpgradePreferred      bool                       `json:"fastUpgradePreferred,omitempty"`
	FinishedStates            string                     `json:"finishedStates,omitempty"`
	IsRUSupported             bool                       `json:"isRuSupported,omitempty"`
	OverallProgress           float64                    `json:"overallProgress"`
	PendingStates             string                     `json:"pendingStates,omitempty"`
	PreviousVersion           string                     `json:"previousVersion,omitempty"`
	RUUnsupportabilityReason  string                     `json:"ruUnsupportabilityReason,omitempty"`
	ScheduleUpgradeAction     string                     `json:"scheduleUpgradeAction,omitempty"`
	ScheduleUpgradeAt         *time.Time                 `json:"scheduleUpgradeAt,omitzero"`
	ScheduleUpgradeMode       string                     `json:"scheduleUpgradeMode,omitempty"`
	StateMachineStatus        string                     `json:"stateMachineStatus,omitempty"`
	StateMachineStatusAt      *time.Time                 `json:"stateMachineStatusAt,omitzero"`
	UpgradeEndAt              *time.Time                 `json:"upgradeEndAt,omitzero"`
	UpgradeEventSeriesID      string                     `json:"upgradeEventSeriesId,omitempty"`
	UpgradeStartAt            *time.Time                 `json:"upgradeStartAt,omitzero"`
	UpgradeStatusV2           *UpgradeStatusV2           `json:"upgradeStatusV2,omitzero"`
	UpgradeRecommendationInfo *UpgradeRecommendationInfo `json:"upgradeRecommendationInfo,omitzero"`
	LastUpgradeDuration       *UpgradeDuration           `json:"lastUpgradeDuration,omitzero"`
}

// IsStaged reports whether the cluster has stagedVersion downloaded and is in
// the ReadyForUpgrade state. Prefers the V2 upgradeStatusV2 payload over the
// legacy V1 fields, falling back to V1 when V2 is absent.
func (info *CDMInfo) IsStaged(stagedVersion string) bool {
	if info == nil {
		return false
	}
	if info.UpgradeStatusV2 != nil {
		return info.UpgradeStatusV2.RSCClusterUpgradeStatus == RSCUpgradeStatusReadyForUpgrade &&
			info.UpgradeStatusV2.UIStatusAttributes.TargetVersion == stagedVersion
	}
	return info.DownloadedVersion == stagedVersion &&
		info.ClusterJobStatus == ClusterJobStatusReadyForUpgrade
}

// UpgradeDetails bundles the cluster identity with its upgrade info, as
// returned by clusterWithUpgradesInfo.
type UpgradeDetails struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	CDMInfo *CDMInfo  `json:"cdmUpgradeInfo,omitzero"`
}
