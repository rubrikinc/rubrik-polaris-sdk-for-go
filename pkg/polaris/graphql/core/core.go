//go:generate go run ../queries_gen.go core

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

// Package core provides a low-level interface to core GraphQL queries provided
// by the Polaris platform. E.g., task chains and enum definitions.
package core

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

type CloudVendor string

const (
	CloudVendorUnknown CloudVendor = "VENDOR_UNKNOWN"
	CloudVendorAWS     CloudVendor = "AWS"
	CloudVendorAzure   CloudVendor = "AZURE"
	CloudVendorGCP     CloudVendor = "GCP"
	CloudVendorOCI     CloudVendor = "OCI"
)

// CloudAccountAction represents a Polaris cloud account action.
type CloudAccountAction string

const (
	Create              CloudAccountAction = "CREATE"
	Delete              CloudAccountAction = "DELETE"
	UpdateChildAccounts CloudAccountAction = "UPDATE_CHILD_ACCOUNTS"
	UpdatePermissions   CloudAccountAction = "UPDATE_PERMISSIONS"
	UpdateRegions       CloudAccountAction = "UPDATE_REGIONS"
)

// PermissionGroup represents a named set of permissions for a feature. Note,
// not all permission groups are applicable to all features.
type PermissionGroup string

const (
	PermissionGroupInvalid                PermissionGroup = "GROUP_UNSPECIFIED"
	PermissionGroupBasic                  PermissionGroup = "BASIC"
	PermissionGroupCCES                   PermissionGroup = "CLOUD_CLUSTER_ES"
	PermissionGroupRSCManagedCluster      PermissionGroup = "RSC_MANAGED_CLUSTER"
	PermissionGroupCustomerManagedCluster PermissionGroup = "CUSTOMER_MANAGED_BASIC"
)

// SortOrder represents the valid sort order values.
type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// Feature represents a Polaris cloud account feature with a set of permission
// groups. If the PermissionGroups field is nil then the full set of permissions
// are used for the feature.
type Feature struct {
	Name             string            `json:"featureType"`
	PermissionGroups []PermissionGroup `json:"permissionsGroups"`
}

// Equal returns true if the features have the same name. Note, this function
// does not compare the permission groups.
func (feature Feature) Equal(other Feature) bool {
	return feature.Name == other.Name
}

// DeepEqual returns true if the features are equal. The features are equal if
// they have the same name and the same permission groups.
func (feature Feature) DeepEqual(other Feature) bool {
	if !feature.Equal(other) {
		return false
	}

	set := make(map[PermissionGroup]struct{}, len(feature.PermissionGroups))
	for _, permissionGroup := range feature.PermissionGroups {
		set[permissionGroup] = struct{}{}
	}
	for _, permissionGroup := range other.PermissionGroups {
		if _, ok := set[permissionGroup]; !ok {
			return false
		}
		delete(set, permissionGroup)
	}

	return len(set) == 0
}

// HasPermissionGroup returns true if the feature has the specified permission
// group.
func (feature Feature) HasPermissionGroup(permissionGroup PermissionGroup) bool {
	return slices.Contains(feature.PermissionGroups, permissionGroup)
}

// String returns a string representation of the feature.
func (feature Feature) String() string {
	if len(feature.PermissionGroups) == 0 {
		return feature.Name
	}

	var buf strings.Builder
	permissionGroups := slices.Clone(feature.PermissionGroups)
	slices.Sort(permissionGroups)
	for _, permissionGroup := range permissionGroups {
		buf.WriteString(string(permissionGroup))
		buf.WriteString(",")
	}

	return fmt.Sprintf("%s(%s)", feature.Name, buf.String()[:buf.Len()-1])
}

// WithPermissionGroups returns a copy of the feature with the specified
// permission groups added.
func (feature Feature) WithPermissionGroups(permissionGroups ...PermissionGroup) Feature {
	groups := append(feature.PermissionGroups, permissionGroups...)
	return Feature{Name: feature.Name, PermissionGroups: groups}
}

var (
	FeatureInvalid                                 = Feature{Name: ""}
	FeatureAll                                     = Feature{Name: "ALL"}
	FeatureAppFlows                                = Feature{Name: "APP_FLOWS"}
	FeatureArchival                                = Feature{Name: "ARCHIVAL"}
	FeatureAzureSQLDBProtection                    = Feature{Name: "AZURE_SQL_DB_PROTECTION"}
	FeatureAzureSQLMIProtection                    = Feature{Name: "AZURE_SQL_MI_PROTECTION"}
	FeatureCloudAccounts                           = Feature{Name: "CLOUDACCOUNTS"} // Deprecated: no replacement.
	FeatureCloudNativeArchival                     = Feature{Name: "CLOUD_NATIVE_ARCHIVAL"}
	FeatureCloudNativeArchivalEncryption           = Feature{Name: "CLOUD_NATIVE_ARCHIVAL_ENCRYPTION"}
	FeatureCloudNativeBlobProtection               = Feature{Name: "CLOUD_NATIVE_BLOB_PROTECTION"}
	FeatureCloudNativeDynamoDBProtection           = Feature{Name: "CLOUD_NATIVE_DYNAMODB_PROTECTION"}
	FeatureCloudNativeProtection                   = Feature{Name: "CLOUD_NATIVE_PROTECTION"}
	FeatureCloudNativeS3Protection                 = Feature{Name: "CLOUD_NATIVE_S3_PROTECTION"}
	FeatureCyberRecoveryDataClassificationData     = Feature{Name: "CYBERRECOVERY_DATA_CLASSIFICATION_DATA"}
	FeatureCyberRecoveryDataClassificationMetadata = Feature{Name: "CYBERRECOVERY_DATA_CLASSIFICATION_METADATA"}
	FeatureDSPMData                                = Feature{Name: "DSPM_DATA"}
	FeatureDSPMMetadata                            = Feature{Name: "DSPM_METADATA"}
	FeatureExocompute                              = Feature{Name: "EXOCOMPUTE"}
	FeatureGCPSharedVPCHost                        = Feature{Name: "GCP_SHARED_VPC_HOST"}
	FeatureKubernetesProtection                    = Feature{Name: "KUBERNETES_PROTECTION"}
	FeatureLaminarCrossAccount                     = Feature{Name: "LAMINAR_CROSS_ACCOUNT"}
	FeatureLaminarInternal                         = Feature{Name: "LAMINAR_INTERNAL"}
	FeatureLaminarOutpostApplication               = Feature{Name: "LAMINAR_OUTPOST_APPLICATION"}
	FeatureLaminarOutpostManagedIdentity           = Feature{Name: "LAMINAR_OUTPOST_MANAGED_IDENTITY"}
	FeatureLaminarTargetApplication                = Feature{Name: "LAMINAR_TARGET_APPLICATION"}
	FeatureLaminarTargetManagedIdentity            = Feature{Name: "LAMINAR_TARGET_MANAGED_IDENTITY"}
	FeatureOutpost                                 = Feature{Name: "OUTPOST"}
	FeatureRDSProtection                           = Feature{Name: "RDS_PROTECTION"}
	FeatureServerAndApps                           = Feature{Name: "SERVERS_AND_APPS"}
)

var validFeatures = map[string]struct{}{
	FeatureAll.Name:                                     {},
	FeatureAppFlows.Name:                                {},
	FeatureArchival.Name:                                {},
	FeatureAzureSQLDBProtection.Name:                    {},
	FeatureAzureSQLMIProtection.Name:                    {},
	FeatureCloudAccounts.Name:                           {}, // Deprecated: no replacement.
	FeatureCloudNativeArchival.Name:                     {},
	FeatureCloudNativeArchivalEncryption.Name:           {},
	FeatureCloudNativeBlobProtection.Name:               {},
	FeatureCloudNativeDynamoDBProtection.Name:           {},
	FeatureCloudNativeProtection.Name:                   {},
	FeatureCloudNativeS3Protection.Name:                 {},
	FeatureCyberRecoveryDataClassificationData.Name:     {},
	FeatureCyberRecoveryDataClassificationMetadata.Name: {},
	FeatureDSPMData.Name:                                {},
	FeatureDSPMMetadata.Name:                            {},
	FeatureExocompute.Name:                              {},
	FeatureGCPSharedVPCHost.Name:                        {},
	FeatureKubernetesProtection.Name:                    {},
	FeatureLaminarCrossAccount.Name:                     {},
	FeatureLaminarInternal.Name:                         {},
	FeatureLaminarOutpostApplication.Name:               {},
	FeatureLaminarOutpostManagedIdentity.Name:           {},
	FeatureLaminarTargetApplication.Name:                {},
	FeatureLaminarTargetManagedIdentity.Name:            {},
	FeatureOutpost.Name:                                 {},
	FeatureRDSProtection.Name:                           {},
	FeatureServerAndApps.Name:                           {},
}

// FeatureNames returns the names of the features.
func FeatureNames(features []Feature) []string {
	var names []string
	for _, feature := range features {
		names = append(names, feature.Name)
	}

	return names
}

// LookupFeature returns the specified feature if it exists in the feature
// slice.
func LookupFeature(features []Feature, feature Feature) (Feature, bool) {
	for _, f := range features {
		if f.Equal(feature) {
			return f, true
		}
	}

	return Feature{}, false
}

// FilterFeaturesOnPermissionGroups verifies that all features either have no
// permission groups or all have permission groups. The features are returned
// in two different slices, depending on whether they have permission groups
// or not.
func FilterFeaturesOnPermissionGroups(features []Feature) ([]string, []Feature, error) {
	if len(features) == 0 {
		return nil, nil, errors.New("no features specified")
	}

	// Check that all features have the same use of permission groups.
	usePG := len(features[0].PermissionGroups) > 0
	for _, feature := range features[1:] {
		if pg := len(feature.PermissionGroups) > 0; pg != usePG {
			return nil, nil, errors.New("features with and without permission groups cannot be mixed")
		}
	}
	if usePG {
		return nil, features, nil
	}

	return FeatureNames(features), nil, nil
}

// Deprecated: use Feature.Name instead.
func FormatFeature(feature Feature) string {
	return strings.ReplaceAll(strings.ToLower(feature.Name), "_", "-")
}

// Deprecated: use Feature{Name: <feature>} instead or ParseFeatureNoValidation
// if you need to remain backwards compatible with previously accepted feature
// names.
func ParseFeature(feature string) (Feature, error) {
	f := ParseFeatureNoValidation(feature)
	if _, ok := validFeatures[f.Name]; ok {
		return f, nil
	}

	return FeatureInvalid, fmt.Errorf("invalid feature: %s", feature)
}

// ParseFeatureNoValidation returns the Feature matching the given feature name.
// No validation is performed.
func ParseFeatureNoValidation(feature string) Feature {
	return Feature{Name: strings.ToUpper(strings.ReplaceAll(feature, "-", "_"))}
}

const (
	// The number of attempts before failing to wait for the Korg job when the
	// error returned is a 403, objects not authorized.
	waitAttempts = 50
)

// Status represents a Polaris cloud account status.
type Status string

const (
	StatusConnected          Status = "CONNECTED"
	StatusConnecting         Status = "CONNECTING"
	StatusDisabled           Status = "DISABLED"
	StatusDisconnected       Status = "DISCONNECTED"
	StatusMissingPermissions Status = "MISSING_PERMISSIONS"
)

// FormatStatus returns the Status as a string using lower-case and with hyphen
// as a separator.
func FormatStatus(status Status) string {
	return strings.ReplaceAll(strings.ToLower(string(status)), "_", "-")
}

// TaskChainState represents the state of a Polaris task chain.
type TaskChainState string

const (
	TaskChainInvalid   TaskChainState = ""
	TaskChainCanceled  TaskChainState = "CANCELED"
	TaskChainCanceling TaskChainState = "CANCELING"
	TaskChainFailed    TaskChainState = "FAILED"
	TaskChainReady     TaskChainState = "READY"
	TaskChainRunning   TaskChainState = "RUNNING"
	TaskChainSucceeded TaskChainState = "SUCCEEDED"
	TaskChainUndoing   TaskChainState = "UNDOING"
)

// API wraps around GraphQL clients to give them the Polaris Core API.
type API struct {
	Version string // Deprecated: use GQL.DeploymentVersion
	GQL     *graphql.Client
	log     log.Logger
}

// Wrap the GraphQL client in the Core API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// TaskChain is a collection of sequential tasks that all must complete for the
// task chain to be considered complete.
type TaskChain struct {
	ID          int64          `json:"id"`
	TaskChainID uuid.UUID      `json:"taskchainUuid"`
	State       TaskChainState `json:"state"`
}

// KorgTaskChainStatus returns the task chain for the specified task chain id.
// If the task chain id refers to a task chain that was just created, its state
// might not have reached ready yet. This can be detected by state being
// TaskChainInvalid and error is nil.
func (a API) KorgTaskChainStatus(ctx context.Context, taskChainID uuid.UUID) (TaskChain, error) {
	a.log.Print(log.Trace)

	query := getKorgTaskchainStatusQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		TaskChainID uuid.UUID `json:"taskchainId,omitempty"`
	}{TaskChainID: taskChainID})
	if err != nil {
		return TaskChain{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Query struct {
				TaskChain TaskChain `json:"taskchain"`
			} `json:"getKorgTaskchainStatus"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return TaskChain{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Query.TaskChain, nil
}

// WaitForTaskChain blocks until the Polaris task chain with the specified task
// chain id has completed. When the task chain completes, the final state of the
// task chain is returned. The wait parameter specifies the amount of time to
// wait before requesting another task status update.
func (a API) WaitForTaskChain(ctx context.Context, taskChainID uuid.UUID, wait time.Duration) (TaskChainState, error) {
	a.log.Print(log.Trace)

	attempt := 0
	for {
		taskChain, err := a.KorgTaskChainStatus(ctx, taskChainID)
		if err != nil {
			var gqlErr graphql.GQLError
			if !errors.As(err, &gqlErr) || len(gqlErr.Errors) < 1 {
				return TaskChainInvalid, fmt.Errorf("failed to get tashchain status for %s: %s", taskChainID, err)
			}
			for _, e := range gqlErr.Errors {
				if e.Extensions.Code == 403 || e.Extensions.Code == 500 {
					continue // Could be a RBAC error that eventually goes away, keep retrying.
				}
				return TaskChainInvalid, fmt.Errorf("unexpected error code when getting tashchain status for %s: %v", taskChainID, err)
			}
			if attempt++; attempt > waitAttempts {
				return TaskChainInvalid, fmt.Errorf("failed to get tashchain status for %s after %d attempts: %s", taskChainID, attempt, err)
			}
			a.log.Printf(log.Debug, "RBAC not ready (attempt: %d)", attempt)
		}

		if taskChain.State == TaskChainSucceeded || taskChain.State == TaskChainCanceled || taskChain.State == TaskChainFailed {
			return taskChain.State, nil
		}

		a.log.Printf(log.Debug, "Waiting for Polaris task chain: %s", taskChainID)

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return TaskChainInvalid, ctx.Err()
		}
	}
}

// WaitForFeatureDisableTaskChain waits for the feature disable task chain to
// finish. If an error occurs while waiting for the task chain or the task chain
// ends in a failed state, an error is returned.
func (a API) WaitForFeatureDisableTaskChain(ctx context.Context, taskChainID uuid.UUID, featureStatus func(ctx context.Context) (bool, error)) error {
	a.log.Print(log.Trace)

	for {
		// Check the status of the task chain.
		taskChain, err := a.KorgTaskChainStatus(ctx, taskChainID)
		if err != nil {
			var gqlErr graphql.GQLError
			if !errors.As(err, &gqlErr) || len(gqlErr.Errors) < 1 {
				return fmt.Errorf("failed to retrieve taskchain status: %s", err)
			}
			for _, e := range gqlErr.Errors {
				if e.Extensions.Code == 403 || e.Extensions.Code == 500 {
					continue // Could be a RBAC error that eventually goes away, keep retrying.
				}
				return fmt.Errorf("unexpected error code when getting tashchain status for %s: %v", taskChainID, err)
			}

			// If the task chain RBAC is not yet ready, we fall back to checking
			// the status of the account feature.
			if disabled, err := featureStatus(ctx); disabled || err != nil {
				return err
			}

			a.log.Printf(log.Debug, "Task chain RBAC not ready")
		} else {
			if taskChain.State == TaskChainSucceeded {
				return nil
			}
			if taskChain.State == TaskChainCanceled || taskChain.State == TaskChainFailed {
				return fmt.Errorf("taskchain failed: task chain state is %s", taskChain.State)
			}
		}

		a.log.Printf(log.Debug, "Waiting for task chain: %s", taskChainID)
		select {
		case <-time.After(10 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Deprecated: use GQL.DeploymentVersion.
func (a API) DeploymentVersion(ctx context.Context) (string, error) {
	a.log.Print(log.Trace)

	query := deploymentVersionQuery
	buf, err := a.GQL.Request(ctx, query, struct{}{})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			DeploymentVersion string `json:"deploymentVersion"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", graphql.UnmarshalError(query, err)
	}

	return payload.Data.DeploymentVersion, nil
}

// DeploymentIPAddresses returns the deployment IP addresses.
func (a API) DeploymentIPAddresses(ctx context.Context) ([]string, error) {
	a.log.Print(log.Trace)

	query := allDeploymentIpAddressesQuery
	buf, err := a.GQL.Request(ctx, query, struct{}{})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			DeploymentIPAddresses []string `json:"allDeploymentIpAddresses"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.DeploymentIPAddresses, nil
}

// EnabledFeaturesForAccount returns all features enable for the RSC account.
func (a API) EnabledFeaturesForAccount(ctx context.Context) ([]Feature, error) {
	a.log.Print(log.Trace)

	query := allEnabledFeaturesForAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct{}{})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Features []string `json:"features"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	var features []Feature
	for _, feature := range payload.Data.Result.Features {
		features = append(features, Feature{Name: feature})
	}
	slices.SortFunc(features, func(i, j Feature) int {
		return cmp.Compare(i.Name, j.Name)
	})

	return features, nil
}

// FormatTimestamp converts a time.Time to RFC3339 format with milliseconds and Z suffix.
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
